package shadowsocksr

import (
    "crypto/hmac"
    "crypto/sha1"
)

func initializeProtocol(ctx *ConnContext) error {
    switch ctx.Protocol {
    case "origin":
        return nil
    case "auth_aes128_md5":
        return initAuthAES128MD5(ctx)
    case "auth_chain_a":
        return initAuthChainA(ctx)
    default:
        return newError("unsupported protocol: " + ctx.Protocol)
    }
}

func initializeObfs(ctx *ConnContext) error {
    switch ctx.Obfs {
    case "plain":
        return nil
    case "http_simple":
        return initHttpSimple(ctx)
    case "tls1.2_ticket_auth":
        return initTLS12TicketAuth(ctx)
    default:
        return newError("unsupported obfs: " + ctx.Obfs)
    }
}

// 实现具体的协议初始化函数
func initAuthAES128MD5(ctx *ConnContext) error {
    // TODO: 实现 auth_aes128_md5 协议初始化
    return nil
}

func initAuthChainA(ctx *ConnContext) error {
    // TODO: 实现 auth_chain_a 协议初始化
    return nil
}

// 实现具体的混淆初始化函数
func initHttpSimple(ctx *ConnContext) error {
    // TODO: 实现 http_simple 混淆初始化
    return nil
}

func initTLS12TicketAuth(ctx *ConnContext) error {
    // TODO: 实现 tls1.2_ticket_auth 混淆初始化
    return nil
}
