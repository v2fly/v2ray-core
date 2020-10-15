<template>
    <div class="inbound">
        <el-row :gutter="10">
            <el-col :span="2" class="op-icons">
                <i type="success" :class="{'el-icon-arrow-down':!showDetail,'el-icon-arrow-up':showDetail}" circle
                   @click="switchInboundDetail"></i>
                <i type="primary" class="el-icon-plus" circle @click="$emit('new-inbound', idx)"></i>
                <i type="primary" class="el-icon-delete" circle @click="$emit('delete-inbound', idx)"></i>
                <i type="primary" class="el-icon-copy-document" circle @click="$emit('copy-inbound', idx, getSettings())"></i>
            </el-col>
            <el-col :span="5">
                <el-input placeholder="tag" v-model="sForm.tag" v-setting/>
            </el-col>
            <el-col :span="5">
                <el-input placeholder="监听地址" v-model="sForm.listen" v-setting/>
            </el-col>
            <el-col :span="6">
                <el-input placeholder="监听端口" v-model="sForm.port" v-setting/>
            </el-col>
            <el-col :span="6">
                <el-select placeholder="协议" v-model="sForm.protocol" @change="changeProtocol" v-setting>
                    <el-option v-for="p of protocols" :label="p" :value="p" :key="p"/>
                </el-select>
            </el-col>
        </el-row>
        <el-row :gutter="10" v-show="showDetail">

            <el-col :span="22" :offset="2">

                <el-card class="box-card">
                    <div slot="header" class="clearfix">
                        <span>{{sForm.protocol}}协议参数</span>
                    </div>
                    <VmessInboundSetting v-model="sForm.settings" v-setting v-if="sForm.protocol=='vmess'"/>
                    <HttpInboundSetting v-model="sForm.settings" v-setting v-if="sForm.protocol=='http'"/>
                    <InboundDokodemoDoorSetting v-model="sForm.settings" v-setting v-if="sForm.protocol=='dokodemo-door'"/>
                    <InboundShadowsocksSetting v-model="sForm.settings" v-setting v-if="sForm.protocol=='shadowsocks'"/>
                    <InboundSocksSetting v-model="sForm.settings" v-setting v-if="sForm.protocol=='socks'"/>
                    <InboundMTProtoSetting v-model="sForm.settings" v-setting v-if="sForm.protocol=='mtproto'"/>
                    <InboundVlessSetting v-model="sForm.settings" v-setting v-if="sForm.protocol=='vless'"/>
                </el-card>
                <el-card class="box-card">
                    <div slot="header" class="clearfix">
                        <span>sniffing</span>
                    </div>
                    <el-form :inline="true" label-width="90px" class="text-left">
                        <el-form-item label="sniffing">
                            <el-switch
                                    v-model="sForm.sniffing.enabled" v-setting>
                            </el-switch>
                        </el-form-item>
                        <el-form-item label="destOverride" label-width="90px">
                            <el-checkbox-group v-model="sForm.sniffing.destOverride" align="left">
                                <el-checkbox label="tls" value="tls" name="destOverride" v-setting></el-checkbox>
                                <el-checkbox label="http" value="http" name="destOverride" v-setting></el-checkbox>
                            </el-checkbox-group>
                        </el-form-item>
                    </el-form>
                </el-card>
                <el-card class="box-card">
                    <div slot="header" class="clearfix">
                        <span>allocate</span>
                    </div>
                    <el-form :inline="true" label-width="90px" class="text-left">
                        <el-form-item label="strategy">
                            <el-tooltip effect="light" slot="label">
                                <div slot="content">端口分配策略。"always" 表示总是分配所有已指定的端口，port 中指定了多少个端口，V2Ray 就会监听这些端口。"random" 表示随机开放端口，每隔 refresh 分钟在 port 范围中随机选取 concurrency 个端口来监听。</div>
                                <label>strategy</label>
                            </el-tooltip>
                            <el-select placeholder="strategy" v-model="sForm.allocate.strategy" v-setting>
                                <el-option label="always" value="always"/>
                                <el-option label="random" value="random"/>
                            </el-select>
                        </el-form-item>
                        <el-form-item label="concurrency">
                            <el-tooltip effect="light" slot="label">
                                <div slot="content">随机端口刷新间隔，单位为分钟。最小值为 2，建议值为 5。这个属性仅当 strategy = random 时有效。</div>
                                <label>concurrency</label>
                            </el-tooltip>
                            <el-input v-model.number="sForm.allocate.concurrency" v-setting/>
                        </el-form-item>
                        <el-form-item label="refresh">
                            <el-tooltip effect="light" slot="label">
                                <div slot="content">随机端口数量。最小值为 1，最大值为 port 范围的三分之一。建议值为 3。</div>
                                <label>refresh</label>
                            </el-tooltip>
                            <el-input v-model.number="sForm.allocate.refresh" v-setting/>
                        </el-form-item>
                    </el-form>
                </el-card>
                <StreamSettingsObject v-model="sForm.streamSettings" v-setting />
            </el-col>

        </el-row>
    </div>
</template>

<script>
    import * as G from '@/consts'
    import VmessInboundSetting from "@/components/inbounds/VmessInboundSetting";
    import HttpInboundSetting from "@/components/inbounds/HttpInboundSetting";
    import InboundDokodemoDoorSetting from "@/components/inbounds/InboundDokodemoDoorSetting";
    import StreamSettingsObject from "@/components/transport/StreamSettingsObject";
    import InboundShadowsocksSetting from "@/components/inbounds/InboundShadowsocksSetting";
    import InboundSocksSetting from "@/components/inbounds/InboundSocksSetting";
    import InboundMTProtoSetting from "@/components/inbounds/InboundMTProtoSetting";
    import InboundVlessSetting from "@/components/inbounds/InboundVlessSetting";

    export default {
        name: "Inbound",
        components: {
            VmessInboundSetting, HttpInboundSetting, InboundDokodemoDoorSetting, StreamSettingsObject,
            InboundShadowsocksSetting, InboundSocksSetting, InboundMTProtoSetting, InboundVlessSetting
        },
        model: {
            prop: 'inbound',
            event: 'change'
        },
        data() {
            return {
                changedByForm: false,
                protocols: G.INBOUND_PROTOCOLS,
                showDetail: false,
                sForm:{
                    "port": 1080,
                    "listen": "0.0.0.0",
                    "protocol": "vmess",
                    "settings": {
                    },
                    "streamSettings": {},
                    "tag": "标识",
                    "sniffing": {
                        "enabled": true,
                        "destOverride": [
                            "http",
                            "tls"
                        ]
                    },
                    "allocate": {
                        "strategy": "always",
                        "refresh": 5,
                        "concurrency": 3
                    }
                }
            }
        },
        created() {
            this.fillDefaultValue(this.inbound);
        },
        mounted() {
            // $(this.$el).on("change", "input", () => {
            //     this.formChanged();
            // });
        },

        watch: {
            inbound: {
                handler: function (val) {
                    if(this.changedByForm){
                        this.changedByForm = false;
                        return;
                    }
                    this.fillDefaultValue(val);
                },
                deep: false
            }
        },
        methods: {
            fillDefaultValue(inbound) {
                inbound = inbound || {};
                Object.assign(this.sForm, inbound);
                this.sForm.sniffing = Object.assign({
                    "enabled": false,
                    "destOverride": []
                }, this.sForm.sniffing||{});
                this.sForm.allocate = Object.assign({
                    "strategy": "always",
                    "refresh": 5,
                    "concurrency": 3
                }, this.sForm.allocate);
                this.$nextTick().then(()=>{
                    this.formChanged();
                });
            },
            formChanged() {
                this.changedByForm = true;
                this.$emit("change", this.getSettings());
            },
            getSettings() {
                let setting = this._.cloneDeep(this.sForm);
                if(!setting.sniffing.enabled){
                    delete setting.sniffing;
                }
                return setting;
            },
            switchInboundDetail() {
                this.showDetail = !this.showDetail;
            },
            changeProtocol() {
                if(this.sForm.protocol === "dokodemo-door"){
                    // 自由门需要禁用sniffing
                    this.sForm.sniffing.enabled = false;
                }

            }
        },
        props: {
            inbound: {
                type: Object,
            },
            idx:{
                type:Number,
                default(){
                    return 0;
                }
            }
        }
    }
</script>

<style scoped>
    .text-left {
        text-align: left;
    }
    .el-select{
        width:100%;
    }
    .inbound{
        padding-bottom: 10px;
    }
    .op-icons{
        line-height: 40px;
    }
</style>
