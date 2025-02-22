package shadowsocksr

import (
    "bytes"
    "crypto/hmac"
    "crypto/rand"
    "crypto/md5"
    "crypto/sha1"
    "encoding/binary"
    "hash/crc32"
    "io"

    "github.com/v2fly/v2ray-core/v5/common"
    "github.com/v2fly/v2ray-core/v5/common/buf"
    "github.com/v2fly/v2ray-core/v5/common/crypto"
    "github.com/v2fly/v2ray-core/v5/common/net"
    "github.com/v2fly/v2ray-core/v5/common/protocol"
    "github.com/v2fly/v2ray-core/v5/common/serial"
)

const (
    Version = 1

    // SSR protocol constants
    AuthKeyLen = 4
    AuthHeaderLen = 7
    AuthDataLen = 7
    ChunkHeaderLen = 2
    ChunkHashLen = 4
    TcpMss = 1452
    UdpMss = 1492

    // Size limits
    ReadBufferSize = 64 * 1024  // 64KB
    WriteBufferSize = 32 * 1024 // 32KB
    MaxPaddingLen = 256

    // Protocol overhead sizes
    ProtocolOverheadSize = 4
    ObfsOverheadSize = 16
)

var addrParser = protocol.NewAddressParser(
    protocol.AddressFamilyByte(0x01, net.AddressFamilyIPv4),
    protocol.AddressFamilyByte(0x04, net.AddressFamilyIPv6),
    protocol.AddressFamilyByte(0x03, net.AddressFamilyDomain),
    protocol.WithAddressTypeParser(func(b byte) byte {
        return b & 0x0F
    }),
)

// SSR conn context stores all the connection state 
type ConnContext struct {
    Protocol      string
    Obfs          string
    ProtocolParam string
    ObfsParam     string
    
    EncryptMethod CipherType
    EncryptKey    []byte
    IV            []byte
    
    ClientID      []byte
    ConnectionID  uint32
    
    TCPMss        int
    
    // Protocol state
    LastClientHash []byte
    LastServerHash []byte
    
    // Data counters
    ReadCount     uint64
    WriteCount    uint64
    
    // Stream ciphers
    Encrypter     cipher.Stream
    Decrypter     cipher.Stream
}

func (c *ConnContext) hmacMD5(key []byte, data []byte) []byte {
    hmacMd5 := hmac.New(md5.New, key)
    hmacMd5.Write(data)
    return hmacMd5.Sum(nil)
}

func (c *ConnContext) hmacSHA1(key []byte, data []byte) []byte {
    hmacSha1 := hmac.New(sha1.New, key)
    hmacSha1.Write(data)
    return hmacSha1.Sum(nil)
}

func (c *ConnContext) InitEncryption() error {
    var err error
    switch c.EncryptMethod {
    case CipherType_AES_128_CFB, CipherType_AES_192_CFB, CipherType_AES_256_CFB:
        block, err := aes.NewCipher(c.EncryptKey)
        if err != nil {
            return newError("failed to create aes cipher").Base(err)
        }
        c.Encrypter = cipher.NewCFBEncrypter(block, c.IV) 
        c.Decrypter = cipher.NewCFBDecrypter(block, c.IV)
        
    case CipherType_CHACHA20:
        c.Encrypter, err = chacha20.NewUnauthenticatedCipher(c.EncryptKey, c.IV)
        if err != nil {
            return newError("failed to create chacha20").Base(err)
        }
        c.Decrypter = c.Encrypter
        
    case CipherType_RC4_MD5:
        md5sum := md5.Sum(append(c.EncryptKey, c.IV...))
        c.Encrypter, err = rc4.NewCipher(md5sum[:])
        if err != nil {
            return newError("failed to create rc4").Base(err)
        }
        c.Decrypter = c.Encrypter
    }
    return nil
}

type TCPReader struct {
    reader  *buf.BufferedReader
    ctx     *ConnContext
    user    *protocol.MemoryUser
    
    buffer  *buf.Buffer
    offset  int
    length  int
}

func NewTCPReader(reader io.Reader, ctx *ConnContext, user *protocol.MemoryUser) *TCPReader {
    return &TCPReader{
        reader: buf.NewBufferedReader(buf.NewReader(reader)),
        ctx:    ctx,
        user:   user,
        buffer: buf.New(),
    }
}

func (r *TCPReader) Read(p []byte) (n int, err error) {
    if r.length == 0 {
        if err := r.fetchChunk(); err != nil {
            return 0, err 
        }
    }

    n = copy(p, r.buffer.BytesFrom(r.offset))
    r.offset += n
    r.length -= n
    return
}

func (r *TCPReader) fetchChunk() error {
    r.buffer.Clear()
    r.offset = 0
    
    // Read chunk size
    size := make([]byte, ChunkHeaderLen) 
    if _, err := io.ReadFull(r.reader, size); err != nil {
        return newError("failed to read chunk size").Base(err)
    }

    length := int(binary.BigEndian.Uint16(size))
    if length > TcpMss {
        return newError("invalid chunk size: ", length)
    }

    // Read chunk data
    if err := r.buffer.Reset(buf.ReadFullFrom(r.reader, int32(length + ChunkHashLen))); err != nil {
        return newError("failed to read chunk data").Base(err)
    }

    // Decrypt chunk
    if r.ctx.Decrypter != nil {
        r.ctx.Decrypter.XORKeyStream(r.buffer.BytesTo(length), r.buffer.BytesTo(length))
    }

    // Verify chunk hash
    chunkHash := sha1.Sum(r.buffer.BytesTo(length))
    if !bytes.Equal(chunkHash[:ChunkHashLen], r.buffer.BytesFrom(length)) {
        return newError("invalid chunk hash")
    }

    r.length = length
    r.ctx.ReadCount += uint64(length)
    
    return nil
}

type TCPWriter struct {
    writer  *buf.BufferedWriter  
    ctx     *ConnContext
    user    *protocol.MemoryUser
}

func NewTCPWriter(writer io.Writer, ctx *ConnContext, user *protocol.MemoryUser) *TCPWriter {
    return &TCPWriter{
        writer: buf.NewBufferedWriter(buf.NewWriter(writer)),
        ctx:    ctx,
        user:   user,
    }
}

func (w *TCPWriter) Write(p []byte) (n int, err error) {
    buffer := buf.New()
    defer buffer.Release()

    for len(p) > 0 {
        chunkSize := TcpMss
        if len(p) < chunkSize {
            chunkSize = len(p)
        }

        // Write chunk header
        binary.BigEndian.PutUint16(buffer.Extend(ChunkHeaderLen), uint16(chunkSize))

        // Write chunk data
        chunk := p[:chunkSize]
        buffer.Write(chunk) 
        
        // Encrypt chunk if needed
        if w.ctx.Encrypter != nil {
            encrypted := buffer.BytesFrom(ChunkHeaderLen)
            w.ctx.Encrypter.XORKeyStream(encrypted, encrypted)
        }

        // Generate and write chunk hash
        chunkHash := sha1.Sum(buffer.BytesRange(ChunkHeaderLen, ChunkHeaderLen+chunkSize))
        buffer.Write(chunkHash[:ChunkHashLen])

        if err := w.writer.WriteMultiBuffer(buf.MultiBuffer{buffer}); err != nil {
            return n, newError("failed to write chunk").Base(err)
        }

        n += chunkSize
        p = p[chunkSize:]
        w.ctx.WriteCount += uint64(chunkSize)
    }

    return n, w.writer.Flush()
}

func ReadTCPSession(user *protocol.MemoryUser, reader io.Reader) (*protocol.RequestHeader, buf.Reader, error) {
    account := user.Account.(*Account)
    ctx := &ConnContext{
        Protocol:      account.Protocol,
        Obfs:         account.Obfs,
        ProtocolParam: account.ProtocolParam,
        ObfsParam:    account.ObfsParam,
        EncryptMethod: account.CipherType,
        EncryptKey:   account.Key,
        TCPMss:      TcpMss,
    }

    buffer := buf.New()
    defer buffer.Release()

    // 1. Read IV
    ivLen := account.Cipher.IVSize()
    if ivLen > 0 {
        if _, err := buffer.ReadFullFrom(reader, ivLen); err != nil {
            return nil, nil, newError("failed to read IV").Base(err)
        }
        ctx.IV = append([]byte(nil), buffer.BytesTo(ivLen)...)
    }

    // 2. Initialize encryption
    if err := ctx.InitEncryption(); err != nil {
        return nil, nil, err
    }

    // 3. Read auth header
    if ctx.Protocol != "origin" {
        if _, err := buffer.ReadFullFrom(reader, AuthHeaderLen); err != nil {
            return nil, nil, newError("failed to read auth header").Base(err)
        }

        // Verify auth header
        data := buffer.Bytes()
        key := append(ctx.EncryptKey, ctx.IV...)
        hmacData := ctx.hmacMD5(key, data[:1])
        if !hmac.Equal(hmacData[:6], data[1:AuthHeaderLen]) {
            return nil, nil, newError("invalid auth header")
        }
    }

    request := &protocol.RequestHeader{
        Version: Version,
        User:    user,
        Command: protocol.RequestCommandTCP,
    }

    // Create TCP reader
    tcpReader := NewTCPReader(reader, ctx, user)

    // Read address
    if addr, port, err := addrParser.ReadAddressPort(buffer, tcpReader); err == nil {
        request.Address = addr
        request.Port = port
    } else {
        return nil, nil, newError("failed to read address").Base(err)
    }

    if request.Address == nil {
        return nil, nil, newError("invalid remote address")
    }

    return request, tcpReader, nil
}

func WriteTCPRequest(request *protocol.RequestHeader, writer io.Writer) (buf.Writer, error) {
    user := request.User
    account := user.Account.(*Account)
    ctx := &ConnContext{
        Protocol:      account.Protocol,
        Obfs:         account.Obfs,
        ProtocolParam: account.ProtocolParam, 
        ObfsParam:    account.ObfsParam,
        EncryptMethod: account.CipherType,
        EncryptKey:   account.Key,
        TCPMss:      TcpMss,
    }

    // 1. Generate and write IV
    if ivLen := account.Cipher.IVSize(); ivLen > 0 {
        iv := make([]byte, ivLen)
        common.Must2(rand.Read(iv))
        if _, err := writer.Write(iv); err != nil {
            return nil, newError("failed to write IV")
        }
        ctx.IV = iv
    }

    // 2. Initialize encryption
    if err := ctx.InitEncryption(); err != nil {
        return nil, err
    }

    // 3. Generate and write auth header
    if ctx.Protocol != "origin" {
        authData := make([]byte, AuthHeaderLen)
        rand.Read(authData[:1])
        key := append(ctx.EncryptKey, ctx.IV...)
        hmacData := ctx.hmacMD5(key, authData[:1])
        copy(authData[1:], hmacData[:6])
        if _, err := writer.Write(authData); err != nil {
            return nil, newError("failed to write auth header")
        }
    }

    tcpWriter := NewTCPWriter(writer, ctx, user)

    // Write address
    if err := addrParser.WriteAddressPort(tcpWriter, request.Address, request.Port); err != nil {
        return nil, newError("failed to write address").Base(err)
    }

    return tcpWriter, nil
}
