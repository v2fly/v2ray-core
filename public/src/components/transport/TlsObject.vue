<template>
    <setting-card title="TLS安全传送配置" :show-enable="true"
                  @update:enableSetting="enableSettingChanged"
                  :enableSetting.sync="enableSetting">
        <template v-slot:header-buttons>
            <i class="el-icon-plus" style="margin-right:10px;" @click="newCert" ></i>
        </template>
        <el-form label-width="100px">
            <el-row >
                <el-col :span="8">
                    <el-form-item label="serverName">
                        <el-tooltip slot="label" effect="light" placement="top">
                            <div slot="content">
                                指定服务器端证书的域名，在连接由 IP 建立时有用。当目标连接由域名指定时，比如在 Socks 入站时接收到了域名，或者由 Sniffing 功能探测出了域名，这个域名会自动用于 serverName，无须手动配置。
                            </div>
                            <label>serverName</label>
                        </el-tooltip>
                        <el-input v-model="sForm.serverName" v-setting/>
                    </el-form-item>
                </el-col>
                <el-col :span="8">
                    <el-form-item label="alpn">
                        <el-tooltip slot="label" effect="light" placement="top">
                            <div slot="content">
                                一个字符串数组，指定了 TLS 握手时指定的 ALPN 数值
                            </div>
                            <label>alpn</label>
                        </el-tooltip>
                        <el-select multiple v-model="sForm.alpn" filterable allow-create v-setting>
                            <el-option value="h2"/>
                            <el-option value="http/1.1"/>
                        </el-select>
                    </el-form-item>
                </el-col>
                <el-col :span="4">
                    <el-form-item label="allowInsecure" label-width="120px">
                        <el-tooltip slot="label" effect="light" placement="top">
                            <div slot="content">
                                是否允许不安全连接（仅用于客户端）。默认值为 false。当值为 true 时，V2Ray 不会检查远端主机所提供的 TLS 证书的有效性。
                            </div>
                            <label>allowInsecure</label>
                        </el-tooltip>
                        <el-switch  v-model="sForm.allowInsecure" v-setting />
                    </el-form-item>
                </el-col>
                <el-col :span="4">
                    <el-form-item label="disableSystemRoot" label-width="130px">
                        <el-tooltip slot="label" effect="light" placement="top">
                            <div slot="content">
                                （V2Ray 4.18+）是否禁用操作系统自带的 CA 证书。默认值为 false。当值为 true 时，V2Ray 只会使用 certificates 中指定的证书进行 TLS 握手。当值为 false 时，V2Ray 只会使用操作系统自带的 CA 证书进行 TLS 握手。
                            </div>
                            <label>disableSystemRoot</label>
                        </el-tooltip>
                        <el-switch  v-model="sForm.disableSystemRoot" v-setting />
                    </el-form-item>
                </el-col>
            </el-row>
        </el-form>
        <CertificateObject v-setting v-for="(c,idx) in sForm.certificates" :key="idx" :idx="idx" :setting="c"
                           @new-cert="newCert"
                           @del-cert="sForm.certificates.splice(idx,1)"
                           @change="sForm.certificates.splice(idx, 1, $event)" />

    </setting-card>
</template>

<script>

    import CertificateObject from "@/components/transport/CertificateObject";
    export default {
        name: "StreamSettingsObject",
        components: {CertificateObject},
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
            newCert(){
                this.sForm.certificates.push({});
            },
            fillDefaultValue(setting) {
                setting = setting || {};
                if (setting.alpn || setting.serverName) {
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
                    "serverName": "v2ray.com",
                    "allowInsecure": false,
                    "alpn": [
                        "h2",
                        "http/1.1"
                    ],
                    "certificates": [{}],
                    "disableSystemRoot": false
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
