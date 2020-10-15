<template>
    <div class="outbound">
        <el-row :gutter="10">
            <el-col :span="2" class="op-icons">
                <i type="success" :class="{'el-icon-arrow-down':!showDetail,'el-icon-arrow-up':showDetail}" circle
                   @click="switchInboundDetail"></i>
                <i type="primary" class="el-icon-plus" circle @click="$emit('new-bound', idx)"></i>
                <i type="primary" class="el-icon-delete" circle @click="$emit('delete-bound', idx)"></i>
                <i type="primary" class="el-icon-copy-document" circle @click="$emit('copy-bound', idx, getSettings())"></i>
                <i type="primary" class="el-icon-upload2" circle @click="$emit('to-top', idx)" v-if="idx>0"></i>
            </el-col>
            <el-col :span="5">
                <el-input placeholder="tag" v-model="sForm.tag" v-setting/>
            </el-col>
            <el-col :span="5">
                <el-input placeholder="发送出口IP" v-model="sForm.sendThrough" v-setting/>
            </el-col>
            <el-col :span="6">
                <el-input placeholder="转发到Tag" v-model="sForm.proxySettings.tag" v-setting/>
            </el-col>
            <el-col :span="6">
                <el-select placeholder="协议" v-model="sForm.protocol" v-setting>
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
                    <VmessOutboundSetting v-model="sForm.settings" v-setting v-if="sForm.protocol=='vmess'"/>
                    <VlessOutboundSetting v-model="sForm.settings" v-setting v-if="sForm.protocol=='vless'"/>
                    <HttpOutboundSetting v-model="sForm.settings" v-setting v-if="sForm.protocol=='http'"/>
                    <OutboundFreedomSetting v-model="sForm.settings" v-setting v-if="sForm.protocol=='freedom'"/>
                    <Blackhole v-model="sForm.settings" v-setting v-if="sForm.protocol=='blackhole'"/>
                    <OutboundDNS v-model="sForm.settings" v-setting v-if="sForm.protocol=='dns'"/>
                    <OutboundEmpty v-model="sForm.settings" v-setting v-if="sForm.protocol=='mtproto'"/>
                    <OutboundShadowsocksSetting v-model="sForm.settings" v-setting v-if="sForm.protocol=='shadowsocks'"/>
                    <OutboundSocksSetting v-model="sForm.settings" v-setting v-if="sForm.protocol=='socks'"/>
                </el-card>
                <el-card class="box-card">
                    <div slot="header" class="clearfix">
                        <span>mux多路复用</span>
                    </div>
                    <el-form :inline="true" label-width="90px" class="text-left">
                        <el-form-item label="mux">
                            <el-switch
                                    v-model="sForm.mux.enabled" v-setting>
                            </el-switch>
                        </el-form-item>
                        <el-form-item label="concurrency">
                            <el-input v-model="sForm.mux.concurrency" v-setting/>
                        </el-form-item>
                    </el-form>
                </el-card>
                <StreamSettingsObject v-model="sForm.streamSettings" v-setting/>
            </el-col>
        </el-row>
    </div>
</template>

<script>
    import * as G from '@/consts'
    import VmessOutboundSetting from "@/components/outbounds/VmessOutboundSetting";
    import HttpOutboundSetting from "@/components/outbounds/HttpOutboundSetting";
    import OutboundFreedomSetting from "@/components/outbounds/OutboundFreedomSetting";
    import StreamSettingsObject from "@/components/transport/StreamSettingsObject";
    import Blackhole from "@/components/outbounds/Blackhole";
    import OutboundDNS from "@/components/outbounds/OutboundDNS";
    import OutboundEmpty from "@/components/outbounds/OutboundEmpty";
    import OutboundShadowsocksSetting from "@/components/outbounds/OutboundShadowsocksSetting";
    import OutboundSocksSetting from "@/components/outbounds/OutboundSocksSetting";
    import VlessOutboundSetting from "@/components/outbounds/VlessOutboundSetting";

    export default {
        name: "Outbound",
        components: {
            VmessOutboundSetting, HttpOutboundSetting, OutboundFreedomSetting,
            StreamSettingsObject, Blackhole, OutboundDNS, OutboundEmpty, OutboundShadowsocksSetting,
            OutboundSocksSetting, VlessOutboundSetting
        },
        model: {
            prop: 'outbound',
            event: 'change'
        },
        data() {
            return {
                changedByForm: false,
                protocols: G.OUTBOUND_PROTOCOLS,
                showDetail: false,
                sForm: {
                    "sendThrough": "0.0.0.0",
                    "protocol": "vmess",
                    "settings": {},
                    "tag": "标识",
                    "streamSettings": {},
                    "proxySettings": {
                        "tag": ""
                    },
                    "mux": {
                        "enabled": false,
                        "concurrency": 8
                    }
                }
            }
        },
        created() {
            this.fillDefaultValue(this.outbound);
        },
        mounted() {
        },

        watch: {
            outbound: {
                handler: function (val) {
                    if (this.changedByForm) {
                        this.changedByForm = false;
                        return;
                    }
                    this.fillDefaultValue(val);
                },
                deep: false
            }
        },
        methods: {
            fillDefaultValue(val){
                val = val || {};
                Object.assign(this.sForm, val);
                this.sForm.mux = Object.assign({
                    "enabled": false,
                    "concurrency": 8
                }, this.sForm.mux);
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
                if(setting.proxySettings.tag == "" ){
                    delete setting.proxySettings;
                }
                if(!setting.mux.enabled){
                    delete setting.mux;
                }
                return setting;
            },
            switchInboundDetail() {
                this.showDetail = !this.showDetail;
            },
        },
        props: {
            outbound: {
                type: Object,
            },
            idx: {
                type: Number,
                default() {
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

    .el-select {
        width: 100%;
    }
    .op-icons{
        line-height: 40px;
    }

    .outbound {
        padding-bottom: 10px;
    }
</style>
