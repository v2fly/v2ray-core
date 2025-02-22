package shadowsocksr

import (
    "crypto/md5"
    "crypto/sha1"
)

func passwordToKey(password string, keyLen int) []byte {
    key := make([]byte, keyLen)
    md5Sum := md5.Sum([]byte(password))
    copy(key, md5Sum[:])

    if keyLen > 16 {
        data := make([]byte, 0, 16+len(password))
        data = append(data, md5Sum[:]...)
        data = append(data, []byte(password)...)
        md5Sum = md5.Sum(data)
        copy(key[16:], md5Sum[:])
    }
    return key
}

func getCipherKeyLen(method CipherType) int {
    switch method {
    case CipherType_AES_128_CFB, CipherType_AES_128_CTR:
        return 16
    case CipherType_AES_192_CFB, CipherType_AES_192_CTR:
        return 24
    case CipherType_AES_256_CFB, CipherType_AES_256_CTR:
        return 32
    case CipherType_CHACHA20, CipherType_CHACHA20_IETF:
        return 32
    case CipherType_RC4_MD5:
        return 16
    default:
        return 0
    }
}

func getCipherIVLen(method CipherType) int {
    switch method {
    case CipherType_AES_128_CFB, CipherType_AES_192_CFB, CipherType_AES_256_CFB:
        return 16
    case CipherType_AES_128_CTR, CipherType_AES_192_CTR, CipherType_AES_256_CTR:
        return 16
    case CipherType_CHACHA20:
        return 8
    case CipherType_CHACHA20_IETF:
        return 12
    case CipherType_RC4_MD5:
        return 16
    default:
        return 0
    }
}
