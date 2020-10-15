<template>
    <setting-card title="mKCP传输数据" :enable-setting.sync="enableSetting" @update:enableSetting="enableSettingChanged">
        <el-form :inline="false" label-width="90px" class="text-left">
            <el-row>
                <el-col :span="8">
                    <el-form-item label="mtu">
                        <template v-slot:label>
                            <el-tooltip placement="top" effect="light">
                                <div slot="content">最大传输单元（maximum transmission unit），请选择一个介于 576 - 1460 之间的值。默认值为 1350。</div>
                                <label>mtu</label>
                            </el-tooltip>
                        </template>
                        <el-input v-model="sForm.mtu" v-setting/>
                    </el-form-item>
                </el-col>
                <el-col :span="8">
                    <el-form-item label="uplinkCapacity" label-width="130px">
                        <template v-slot:label>
                            <el-tooltip placement="top" effect="light">
                                <div slot="content">上行链路容量，即主机发出数据所用的最大带宽，单位 MB/s，默认值 5。注意是 Byte 而非 bit。可以设置为 0，表示一个非常小的带宽。</div>
                                <label>uplinkCapacity</label>
                            </el-tooltip>
                        </template>
                        <el-input v-model="sForm.uplinkCapacity" v-setting/>
                    </el-form-item>
                </el-col>
                <el-col :span="8">
                    <el-form-item label="readBufferSize" label-width="120px">
                        <template v-slot:label>
                            <el-tooltip placement="top" effect="light">
                                <div slot="content">单个连接的读取缓冲区大小，单位是 MB。默认值为 2。</div>
                                <label>readBufferSize</label>
                            </el-tooltip>
                        </template>
                        <el-input v-model="sForm.readBufferSize" v-setting/>
                    </el-form-item>
                </el-col>
            </el-row>
            <el-row>
                <el-col :span="8">
                    <el-form-item label="tti">
                        <template v-slot:label>
                            <el-tooltip placement="top" effect="light">
                                <div slot="content">传输时间间隔（transmission time interval），单位毫秒（ms），mKCP 将以这个时间频率发送数据。请选译一个介于 10 - 100 之间的值。默认值为 50。</div>
                                <label>tti</label>
                            </el-tooltip>
                        </template>
                        <el-input v-model="sForm.tti" v-setting/>
                    </el-form-item>
                </el-col>
                <el-col :span="8">
                    <el-form-item label="downlinkCapacity" label-width="130px">
                        <template v-slot:label>
                            <el-tooltip placement="top" effect="light">
                                <div slot="content">下行链路容量，即主机接收数据所用的最大带宽，单位 MB/s，默认值 20。注意是 Byte 而非 bit。可以设置为 0，表示一个非常小的带宽。</div>
                                <label>downlinkCapacity</label>
                            </el-tooltip>
                        </template>
                        <el-input v-model="sForm.downlinkCapacity" v-setting/>
                    </el-form-item>
                </el-col>
                <el-col :span="8">
                    <el-form-item label="writeBufferSize" label-width="120px">
                        <template v-slot:label>
                            <el-tooltip placement="top" effect="light">
                                <div slot="content">单个连接的写入缓冲区大小，单位是 MB。默认值为 2。</div>
                                <label>writeBufferSize</label>
                            </el-tooltip>
                        </template>
                        <el-input v-model="sForm.writeBufferSize" v-setting/>
                    </el-form-item>
                </el-col>
            </el-row>
            <el-row>
                <el-col :span="8">
                    <el-form-item label="seed">
                        <template v-slot:label>
                            <el-tooltip placement="top" effect="light">
                                <div slot="content">v4.24.2+，可选的混淆密码，使用 AES-128-GCM 算法混淆流量数据，客户端和服务端需要保持一致，启用后会输出"NewAEADAESGCMBasedOnSeed Used"到命令行。本混淆机制不能用于保证通信内容的安全，但可能可以对抗部分封锁，在开发者测试环境下开启此设置后没有出现原版未混淆版本的封端口现象。</div>
                                <label>seed</label>
                            </el-tooltip>
                        </template>
                        <el-input v-model="sForm.seed" v-setting/>
                    </el-form-item>
                </el-col>

                <el-col :span="8">
                    <el-form-item label="头部伪装" label-width="130px">
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
                </el-col>
                <el-col :span="8">
                    <el-form-item label="congestion" label-width="120px">
                        <template v-slot:label>
                            <el-tooltip placement="top" effect="light">
                                <div slot="content">是否启用拥塞控制，默认值为 false。开启拥塞控制之后，V2Ray 会自动监测网络质量，当丢包严重时，会自动降低吞吐量；当网络畅通时，也会适当增加吞吐量。</div>
                                <label>congestion</label>
                            </el-tooltip>
                        </template>
                        <el-switch v-model="sForm.congestion" v-setting/>
                    </el-form-item>
                </el-col>
            </el-row>

        </el-form>

        <p>mKCP 使用 UDP 来模拟 TCP 连接，请确定主机上的防火墙配置正确。mKCP 牺牲带宽来降低延迟。传输同样的内容，mKCP 一般比 TCP 消耗更多的流量</p>

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
                    "mtu": 1350,
                    "tti": 20,
                    "uplinkCapacity": 5,
                    "downlinkCapacity": 20,
                    "congestion": false,
                    "readBufferSize": 1,
                    "writeBufferSize": 1,
                    "header": {
                        "type": "none"
                    },
                    "seed": ""
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
