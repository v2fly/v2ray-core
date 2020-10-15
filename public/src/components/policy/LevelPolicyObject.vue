<template>
    <setting-card :title="'用户等级策略'+level" :show-enable="false"
                  @update:enableSetting="enableSettingChanged"
                  :enableSetting.sync="enableSetting">
        <template v-slot:header-buttons>
            <i class="el-icon-delete" style="margin-right:10px;" @click="$emit('del-level')"></i>
        </template>
        <el-form label-width="100px" label-position="right">
            <el-row >
                <el-col :span="8">
                    <el-form-item label="handshake">
                        <el-tooltip slot="label" effect="light" placement="top">
                            <div slot="content">
                                <p>连接建立时的握手时间限制。单位为秒。默认值为 4。在入站代理处理一个新连接时，在握手阶段（比如 VMess
                                    读取头部数据，判断目标服务器地址），如果使用的时间超过这个时间，则中断该连接。</p>
                            </div>
                            <label>handshake</label>
                        </el-tooltip>
                        <el-input v-model.number="sForm.handshake" v-setting/>
                    </el-form-item>
                </el-col>
                <el-col :span="8">
                    <el-form-item label="connIdle">
                        <el-tooltip slot="label" effect="light" placement="top">
                            <div slot="content">
                                连接空闲的时间限制。单位为秒。默认值为 300。在入站出站代理处理一个连接时，如果在 connIdle 时间内，没有任何数据被传输（包括上行和下行数据），则中断该连接。
                            </div>
                            <label>connIdle</label>
                        </el-tooltip>
                        <el-input v-model.number="sForm.connIdle" v-setting/>
                    </el-form-item>
                </el-col>
                <el-col :span="8">
                    <el-form-item label="bufferSize">
                        <el-tooltip slot="label" effect="light" placement="top">
                            <div slot="content">
                                每个连接的内部缓存大小。单位为 kB。当值为 0 时，内部缓存被禁用。
                                <p>默认值 (V2Ray 4.4+):</p>
                                <ul>
                                    <li>在 ARM、MIPS、MIPSLE 平台上，默认值为 <code>0</code>。</li>
                                    <li>在 ARM64、MIPS64、MIPS64LE 平台上，默认值为 <code>4</code>。</li>
                                    <li>在其它平台上，默认值为 <code>512</code>。</li>
                                </ul>
                                <p>默认值 (V2Ray 4.3-):</p>
                                <ul>
                                    <li>在 ARM、MIPS、MIPSLE、ARM64、MIPS64、MIPS64LE 平台上，默认值为 <code>16</code>。</li>
                                    <li>在其它平台上，默认值为 <code>2048</code>。</li>
                                </ul>
                            </div>
                            <label>bufferSize</label>
                        </el-tooltip>
                        <el-input v-model.number="sForm.bufferSize" v-setting/>
                    </el-form-item>
                </el-col>

                <el-col :span="8">
                    <el-form-item label="uplinkOnly">
                        <el-tooltip slot="label" effect="light" placement="top">
                            <div slot="content">
                                当连接下行线路关闭后的时间限制。单位为秒。默认值为 2。当服务器（如远端网站）关闭下行连接时，出站代理会在等待 uplinkOnly 时间后中断连接。
                            </div>
                            <label>uplinkOnly</label>
                        </el-tooltip>
                        <el-input v-model.number="sForm.uplinkOnly" v-setting/>
                    </el-form-item>
                </el-col>

                <el-col :span="8">
                    <el-form-item label="downlinkOnly">
                        <el-tooltip slot="label" effect="light" placement="top">
                            <div slot="content">
                                当连接上行线路关闭后的时间限制。单位为秒。默认值为 5。当客户端（如浏览器）关闭上行连接时，入站代理会在等待 downlinkOnly 时间后中断连接。
                            </div>
                            <label>downlinkOnly</label>
                        </el-tooltip>
                        <el-input v-model.number="sForm.downlinkOnly" v-setting/>
                    </el-form-item>
                </el-col>

                <el-col :span="4">
                    <el-form-item label="statsUserUplink" label-width="140px">
                        <el-tooltip slot="label" effect="light" placement="top">
                            <div slot="content">
                                当值为 true 时，开启当前等级的所有用户的上行流量统计。
                            </div>
                            <label>statsUserUplink</label>
                        </el-tooltip>
                        <el-switch v-model="sForm.statsUserUplink" v-setting/>
                    </el-form-item>
                </el-col>
                <el-col :span="4">
                    <el-form-item label="statsUserDownlink" label-width="140px">
                        <el-tooltip slot="label" effect="light" placement="top">
                            <div slot="content">
                                当值为 true 时，开启当前等级的所有用户的下行流量统计。
                            </div>
                            <label>statsUserDownlink</label>
                        </el-tooltip>
                        <el-switch v-model="sForm.statsUserDownlink" v-setting/>
                    </el-form-item>
                </el-col>



            </el-row>
        </el-form>

    </setting-card>
</template>

<script>


    export default {
        name: "LevelPolicyObject",
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
                Object.assign(this.sForm, setting);
                this.$nextTick(() => {
                    this.formChanged();
                });
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
                let setting = this.getSettings();
                if(this.setting == null && setting == null){
                    return;
                }
                this.changedByForm = true;
                this.$emit("change", setting);
            }
        },
        computed: {},
        data() {
            return {
                sForm: {
                    "handshake": 4,
                    "connIdle": 300,
                    "uplinkOnly": 2,
                    "downlinkOnly": 5,
                    "statsUserUplink": false,
                    "statsUserDownlink": false,
                    "bufferSize": 10240
                },
                changedByForm: false,
                "enableSetting": true,
            }
        },
        props: {
            setting: {
                type: Object
            },
            level: {
                type: String,
                default() {
                    return 0
                }
            }
        }
    }
</script>

<style>
    .el-select {
        width: 100%;
    }

    .el-form--label-top .el-form-item__label {
        line-height: normal;
        padding: 0;
    }
</style>
