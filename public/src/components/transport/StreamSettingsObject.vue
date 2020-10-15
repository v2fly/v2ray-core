<template>
    <setting-card title="底层传输配置" :show-enable="true"
                  @update:enableSetting="enableSettingChanged"
                  :enableSetting.sync="enableSetting">

        <el-form label-width="100px">
            <el-row :gutter="10">
                <el-col :span="8">
                    <el-form-item label="network">
                        <el-select v-model="sForm.network" v-setting>
                            <el-option value="tcp"/>
                            <el-option value="kcp"/>
                            <el-option value="ws"/>
                            <el-option value="http"/>
                            <el-option value="domainsocket"/>
                            <el-option value="quic"/>
                        </el-select>
                    </el-form-item>
                </el-col>
                <el-col :span="8">
                    <el-form-item label="security">
                        <el-select v-model="sForm.security" v-setting>
                            <el-option value="none"/>
                            <el-option value="tls"/>
                        </el-select>
                    </el-form-item>
                </el-col>
            </el-row>
            <el-row :gutter="10">
                <el-col :span="8">
                    <el-form-item label="mark">
                        <el-tooltip slot="label" effect="light" placement="top">
                            <div slot="content">
                                <p>一个整数。当其值非零时，在出站连接上标记 SO_MARK。</p>
                                <ul>
                                    <li>仅适用于 Linux 系统。</li>
                                    <li>需要 CAP_NET_ADMIN 权限。</li>
                                </ul>
                            </div>
                            <label>mark</label>
                        </el-tooltip>
                        <el-input v-model="sForm.sockopt.mark" v-setting/>
                    </el-form-item>
                </el-col>
                <el-col :span="8">
                    <el-form-item label="tproxy">
                        <el-tooltip slot="label" effect="light" placement="top">
                            <div slot="content">
                                <p>是否开启透明代理（仅适用于 Linux）。</p>
                                <ul>
                                    <li><code>"redirect"</code>：使用 Redirect 模式的透明代理。仅支持 TCP/IPv4 和 UDP 连接。</li>
                                    <li><code>"tproxy"</code>：使用 TProxy 模式的透明代理。支持 TCP 和 UDP 连接。</li>
                                    <li><code>"off"</code>：关闭透明代理。</li>
                                </ul>
                                <p>透明代理需要 Root 或 CAP_NET_ADMIN 权限。</p>
                            </div>
                            <label>tproxy</label>
                        </el-tooltip>
                        <el-select v-model="sForm.sockopt.tproxy" v-setting>
                            <el-option value="redirect"/>
                            <el-option value="tproxy"/>
                            <el-option value="off"/>
                        </el-select>
                    </el-form-item>
                </el-col>
                <el-col :span="8">
                    <el-form-item label="tcpFastOpen" label-width="120px">
                        <el-tooltip slot="label" effect="light" placement="top">
                            <div slot="content">
                                <p>是否启用 <a href="https://zh.wikipedia.org/wiki/TCP%E5%BF%AB%E9%80%9F%E6%89%93%E5%BC%80"
                                           target="_blank" rel="noopener noreferrer">TCP Fast Open
                                    <svg xmlns="http://www.w3.org/2000/svg" aria-hidden="true" x="0px" y="0px"
                                         viewBox="0 0 100 100" width="15" height="15" class="icon outbound">
                                        <path fill="currentColor"
                                              d="M18.8,85.1h56l0,0c2.2,0,4-1.8,4-4v-32h-8v28h-48v-48h28v-8h-32l0,0c-2.2,0-4,1.8-4,4v56C14.8,83.3,16.6,85.1,18.8,85.1z"></path>
                                        <polygon fill="currentColor"
                                                 points="45.7,48.7 51.3,54.3 77.2,28.5 77.2,37.2 85.2,37.2 85.2,14.9 62.8,14.9 62.8,22.9 71.5,22.9"></polygon>
                                    </svg>
                                </a>。当其值为 <code>true</code> 时，强制开启 TFO；当其值为 <code>false</code> 时，强制关闭
                                    TFO；当此项不存在时，使用系统默认设置。可用于入站出站连接。
                                </p>
                                <ul>
                                    <li>仅在以下版本（或更新版本）的操作系统中可用:
                                        <ul>
                                            <li>Windows 10 (1604)</li>
                                            <li>Mac OS 10.11 / iOS 9</li>
                                            <li>Linux 3.16：系统已默认开启，无需配置。</li>
                                        </ul>
                                    </li>
                                </ul>
                            </div>
                            <label>tcpFastOpen</label>
                        </el-tooltip>
                        <el-switch v-model="sForm.sockopt.tcpFastOpen" v-setting/>
                    </el-form-item>
                </el-col>
            </el-row>
        </el-form>
        <TlsObject v-setting v-model="sForm.tlsSettings" v-if="sForm.security === 'tls'" />
        <TcpObject v-setting v-model="sForm.tcpSettings" v-if="sForm.network==='tcp'"/>
        <KcpObject v-setting v-model="sForm.kcpSettings" v-if="sForm.network==='kcp'"/>
        <WebSocketObject v-setting v-model="sForm.wsSettings" v-if="sForm.network==='ws'"/>
        <Http2Object v-setting v-model="sForm.httpSettings" v-if="sForm.network==='http'"/>
        <DomainSocketObject v-setting v-model="sForm.dsSettings" v-if="sForm.network==='domainsocket'"/>
        <QuicObject v-setting v-model="sForm.quicSettings" v-if="sForm.network==='quic'"/>
    </setting-card>
</template>

<script>
    import WebSocketObject from "@/components/transport/WebSocketObject";
    import Http2Object from "@/components/transport/Http2Object";
    import DomainSocketObject from "@/components/transport/DomainSocketObject";
    import TcpObject from "@/components/transport/TcpObject";
    import QuicObject from "@/components/transport/QuicObject";
    import KcpObject from "@/components/transport/KcpObject";
    import TlsObject from "@/components/transport/TlsObject";

    export default {
        name: "StreamSettingsObject",
        components: {WebSocketObject, Http2Object, DomainSocketObject, TcpObject, QuicObject, KcpObject, TlsObject},
        model: {
            prop: 'setting',
            event: 'change'
        },
        created() {
            this.fillDefaultValue(this.setting);
        },
        mounted() {
        },
        watch: {
            setting(val) {
                if (this.changedByForm) {
                    this.changedByForm = false;
                    return;
                }
                this.fillDefaultValue(val);
            }
        },
        methods: {
            fillDefaultValue(setting) {
                setting = setting || {};
                if (setting.network || setting.security) {
                    this.enableSetting = true;
                } else {
                    this.enableSetting = false;
                }
                Object.assign(this.sForm, setting);
                this.formChanged();
            },
            getSettings() {
                if (!this.enableSetting) {
                    return null;
                }
                return Object.assign({}, this.sForm);
            },
            enableSettingChanged() {
                this.$nextTick(() => {
                    this.formChanged();
                });

            },
            formChanged() {

                this.changedByForm = true;
                this.$emit("change", this.getSettings());
            }
        },
        data() {
            return {
                sForm: {
                    "network": "tcp",
                    "security": "none",
                    "tlsSettings": {},
                    "tcpSettings": {},
                    "kcpSettings": {},
                    "wsSettings": {},
                    "httpSettings": {},
                    "dsSettings": {},
                    "quicSettings": {},
                    "sockopt": {
                        "mark": 0,
                        "tcpFastOpen": false,
                        "tproxy": "off"
                    }
                },
                changedByForm: false,
                "enableSetting": false,
            }
        },
        props: {
            setting: {
                type: Object
            }
        }
    }
</script>

<style scoped>
    .el-select {
        width: 100%;
    }
</style>
