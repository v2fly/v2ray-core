<template>
    <setting-card title="QUIC传输数据" :enable-setting.sync="enableSetting" @update:enableSetting="enableSettingChanged">
        <el-form :inline="true" label-width="90px" class="text-left">
            <el-form-item label="security">
                <template v-slot:label>
                    <el-tooltip placement="top" effect="light">
                        <div slot="content">此加密是对 QUIC 数据包的加密，加密后数据包无法被探测。</div>
                        <label>security</label>
                    </el-tooltip>
                </template>
                <el-select v-model="sForm.security" v-setting>
                    <el-option value="none"/>
                    <el-option value="aes-128-gcm"/>
                    <el-option value="chacha20-poly1305"/>
                </el-select>
            </el-form-item>
            <el-form-item label="key">
                <template v-slot:label>
                    <el-tooltip placement="top" effect="light">
                        <div slot="content">加密时所用的密钥。可以是任意字符串。当 security 不为 "none" 时有效。</div>
                        <label>key</label>
                    </el-tooltip>
                </template>
                <el-input v-model="sForm.key" v-setting/>
            </el-form-item>
            <el-form-item label="头部伪装">
                <template v-slot:label>
                    <el-tooltip placement="top" effect="light">
                        <div slot="content">
                            <ul>
                                <li><code>"none"</code>：默认值，不进行伪装，发送的数据是没有特征的数据包。</li>
                                <li><code>"srtp"</code>：伪装成 SRTP 数据包，会被识别为视频通话数据（如 FaceTime）。</li>
                                <li><code>"utp"</code>：伪装成 uTP 数据包，会被识别为 BT 下载数据。</li>
                                <li><code>"wechat-video"</code>：伪装成微信视频通话的数据包。</li>
                                <li><code>"dtls"</code>：伪装成 DTLS 1.2 数据包。</li>
                                <li><code>"wireguard"</code>：伪装成 WireGuard 数据包。（并不是真正的 WireGuard 协议）</li>
                            </ul>
                        </div>
                        <label>头部伪装</label>
                    </el-tooltip>
                </template>
                <el-select v-model="sForm.header.type" v-setting>
                    <el-option value="none"/>
                    <el-option value="srtp"/>
                    <el-option value="utp"/>
                    <el-option value="wechat-video"/>
                    <el-option value="dtls"/>
                    <el-option value="wireguard"/>
                </el-select>
            </el-form-item>
        </el-form>
        <p>QUIC 全称 Quick UDP Internet Connection，是由 Google 提出的使用 UDP 进行多路并发传输的协议。其主要优势是:</p>
        <ol>
            <li>减少了握手的延迟（1-RTT 或 0-RTT）</li>
            <li>多路复用，并且没有 TCP 的阻塞问题</li>
            <li>连接迁移，（主要是在客户端）当由 Wifi 转移到 4G 时，连接不会被断开。</li>
        </ol>
        <p>QUIC 目前处于实验期，使用了正在标准化过程中的 IETF 实现，不能保证与最终版本的兼容性。</p>
    </setting-card>
</template>

<script>
    export default {
        name: "DomainSocketObject",
        model: {
            prop: 'setting',
            event: 'change'
        },
        data() {
            return {
                enableSetting: false,
                changedByForm: false,
                sForm: {
                    "security": "none",
                    "key": "",
                    "header": {
                        "type": "none"
                    }
                }

            }
        },
        watch: {
            setting: {
                handler(val) {
                    if (this.changedByForm) {
                        this.changedByForm = false;
                        return;
                    }
                    this.fillDefaultValue(val);
                },
                deep: false
            }
        },
        created() {
            let setting = this.setting || {};
            this.fillDefaultValue(setting);

        },
        mounted() {
        },
        methods: {
            fillDefaultValue(setting) {
                setting = this.setting || {};
                if (this._.isEmpty(setting)) {
                    this.enableSetting = false;
                } else {
                    this.enableSetting = true;
                }
                Object.assign(this.sForm, setting);

                this.formChanged();
            },

            getSettings() {
                if (!this.enableSetting) {
                    return null;
                }
                let setting = Object.assign({}, this.sForm);
                return setting;
            },
            formChanged() {
                this.changedByForm = true;
                this.$emit("change", this.getSettings());
            },
            enableSettingChanged() {
                this.$nextTick(() => {
                    this.formChanged();
                });

            },
        },
        props: {
            setting: {
                type: Object,
            }
        }
    }
</script>

<style scoped>

</style>
