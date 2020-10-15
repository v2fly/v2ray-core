<template>
    <setting-card :title="'证书详情'+(idx+1)+'-'+sForm.usage" :show-enable="false"
                  @update:enableSetting="enableSettingChanged"
                  :enableSetting.sync="enableSetting">
        <template v-slot:header-buttons>
            <i class="el-icon-plus" style="margin-right:10px;" @click="$emit('new-cert')" ></i>
            <i class="el-icon-delete" style="margin-right:10px;" @click="$emit('del-cert')"></i>
        </template>
        <el-form label-width="100px" label-position="top">
            <el-row :gutter="10">
                <el-col :span="8">
                    <el-form-item label="usage">
                        <el-tooltip slot="label" effect="light" placement="top">
                            <div slot="content">
                                <ul>
                                    <li><code>"encipherment"</code>：证书用于 TLS 认证和加密。</li>
                                    <li><code>"verify"</code>：证书用于验证远端 TLS 的证书。当使用此项时，当前证书必须为 CA 证书。</li>
                                    <li><code>"issue"</code>：证书用于签发其它证书。当使用此项时，当前证书必须为 CA 证书。</li>
                                </ul>
                            </div>
                            <label>usage</label>
                        </el-tooltip>
                        <el-select v-model="sForm.usage" v-setting>
                            <el-option value="encipherment"/>
                            <el-option value="verify"/>
                            <el-option value="issue"/>
                        </el-select>
                    </el-form-item>
                </el-col>

                <el-col :span="8">
                    <el-form-item label="certificateFile">
                        <el-tooltip slot="label" effect="light" placement="top">
                            <div slot="content">
                                <p>证书文件路径，如使用 OpenSSL 生成，后缀名为 .crt。</p>
                                <p>使用 v2ctl cert -ca 可以生成自签名的 CA 证书。</p>
                            </div>
                            <label>certificateFile</label>
                        </el-tooltip>
                        <el-input v-model="sForm.certificateFile" v-setting/>
                    </el-form-item>
                </el-col>
                <el-col :span="8">
                    <el-form-item label="keyFile">
                        <el-tooltip slot="label" effect="light" placement="top">
                            <div slot="content">
                                <p>密钥文件路径，如使用 OpenSSL 生成，后缀名为 .key。目前暂不支持需要密码的 key 文件。</p>
                            </div>
                            <label>keyFile</label>
                        </el-tooltip>
                        <el-input v-model="sForm.keyFile" v-setting/>
                    </el-form-item>
                </el-col>

                <el-col :span="12">
                    <el-form-item label="certificate">
                        <el-tooltip slot="label" effect="light" placement="top">
                            <div slot="content">
                                <p>一个字符串数组，表示证书内容，格式如样例所示。certificate 和 certificateFile 二者选一。</p>
                            </div>
                            <label>certificate</label>
                        </el-tooltip>
                        <el-input type="textarea" rows="8" v-model="sForm.certificateValue" v-setting/>
                    </el-form-item>
                </el-col>
                <el-col :span="12">
                    <el-form-item label="key">
                        <el-tooltip slot="label" effect="light" placement="top">
                            <div slot="content">
                                <p>一个字符串数组，表示密钥内容，格式如样例如示。key 和 keyFile 二者选一。</p>
                            </div>
                            <label>key</label>
                        </el-tooltip>
                        <el-input type="textarea" rows="8" v-model="sForm.keyValue" v-setting/>
                    </el-form-item>
                </el-col>

            </el-row>
        </el-form>

    </setting-card>
</template>

<script>


    export default {
        name: "CertificateObject",
        components: {},
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
        computed: {
            keyValue: {
                get() {
                    if (!this.sForm.key) {
                        return "";
                    }
                    return this.sForm.key.join("\n");
                },
                set(val) {
                    this.sForm.key = val.split("\n");
                }
            },
            certificateValue: {
                get() {
                    if (!this.sForm.certificate) {
                        return "";
                    }
                    return this.sForm.certificate.join("\n");
                },
                set(val) {
                    this.sForm.certificate = val.split("\n");
                }
            },
        },
        data() {
            return {
                sForm: {
                    "usage": "encipherment",
                    "certificateFile": "",
                    "keyFile": "",
                    "certificate": [],
                    "key": []
                },
                changedByForm: false,
                "enableSetting": false,
            }
        },
        props: {
            setting: {
                type: Object
            },
            idx: {
                type: Number,
                default() {
                    return 1
                }
            }
        }
    }
</script>

<style>
    .el-select {
        width: 100%;
    }
    .el-form--label-top .el-form-item__label{
        line-height: normal;
        padding:0;
    }
</style>
