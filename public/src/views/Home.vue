<template>
    <div class="home">
        <el-row>
            <el-col>
                <el-form :inline="true" >
                    <el-form-item label="API接口地址:">
                        <el-input v-model="baseURL" style="width:300px" placeholder="api 接口地址"/>
                    </el-form-item>
                    <el-form-item>
                        <el-button @click="updateBaseURL">确定</el-button>
                    </el-form-item>
                </el-form>
            </el-col>
        </el-row>
        <el-row>
            <el-col>
                <el-button @click="loadSetting">加载</el-button>
                <el-button @click="showSetting">查看</el-button>
                <el-button @click="saveSetting">保存</el-button>
                <el-button @click="reloadServer">reload</el-button>
                <el-button-group>
                    <el-button @click="foldAllSetting" icon="el-icon-arrow-up">折叠</el-button>
                    <el-button @click="unfoldAllSetting" icon="el-icon-arrow-down">展开</el-button>
                </el-button-group>

            </el-col>
        </el-row>

        <LogSetting v-model="sForm.log" @change="logChanged" data-setting="1"/>
        <ApiSetting v-model="sForm.api" data-setting="1"/>
        <AdminSetting v-model="sForm.admin" data-setting="1"/>
        <DnsSetting v-model="sForm.dns" data-setting="1"/>
        <PolicyObject v-model="sForm.policy" data-setting="1"/>
        <Reverse v-model="sForm.reverse" data-setting="1"/>
        <Inbounds v-model="sForm.inbounds" data-setting="1"/>
        <Outbounds v-model="sForm.outbounds" data-setting="1"/>
        <Routing v-model="sForm.routing" data-setting="1"/>
        <GlobalTransport v-model="sForm.transport" data-setting="1"/>
        <el-dialog :visible.sync="dialogVisible" title="配置信息" center width="100%" fullscreen >
            <el-input type="textarea" placeholder="placeholder" rows="10" :autosize='{minRows:10, maxRows:30}' v-model="configJson"></el-input>
            <div slot="footer" class="dialog-footer">
                <el-button @click="dialogVisible = false">关闭</el-button>
                <el-button type="primary" @click="updateSetting">更新配置</el-button>
            </div>
        </el-dialog>
    </div>
</template>

<script>
    // @ is an alias to /src
    import LogSetting from "@/components/LogSetting";
    import ApiSetting from "@/components/ApiSetting";
    import Inbounds from "@/components/inbounds/Inbounds";
    import Outbounds from "@/components/outbounds/Outbounds";
    import SettingCard from "@/components/SettingCard";
    import Routing from "@/components/routing/Routing";
    import Reverse from "@/components/reverse/Reverse";
    import Vue from 'vue';
    import GlobalTransport from "@/components/transport/GlobalTransport";
    import PolicyObject from "@/components/policy/PolicyObject";
    import DnsSetting from "@/components/dns/DnsSetting";
    import AdminSetting from "@/components/AdminSetting";
    import Config from "@/api/Config";
    import ajax from "@/api/lib/ajax";

    Vue.component(SettingCard.name, SettingCard);

    Vue.directive("setting", {
        bind(el, binding, vnode) {
            if (vnode.componentInstance) {
                vnode.componentInstance.$on("change", () => {
                    if (vnode.context.formChanged) {
                        vnode.context.formChanged();
                    }
                });
            } else if (vnode.elm) {
                vnode.elm.onchange = () => {
                    if (vnode.context.formChanged) {
                        vnode.context.formChanged();
                    }
                };
            }

        }
    });

    export default {
        name: 'Home',
        data() {
            return {
                dialogVisible: false,
                configJson: "",
                baseURL:"http://localhost:8089/v2ray",
                sForm: {
                    log: {
                        "access": "",
                        "error": "",
                        "loglevel": "warning"
                    },
                    api: {
                        "tag": "api",
                        "services": [
                            "HandlerService",
                            "LoggerService",
                            "StatsService"
                        ]
                    },
                    admin: null,
                    dns: null,
                    inbounds: [],
                    outbounds: [],
                    routing: {},
                    policy: null,
                    reverse: null,
                    transport: null,
                }

            }
        },
        created() {
            this.loadSetting();
            this.baseURL = ajax.getBaseURL();
        },
        components: {
            LogSetting, ApiSetting, Inbounds, Outbounds, Routing, Reverse,
            GlobalTransport, PolicyObject, DnsSetting, AdminSetting
        },
        methods: {
            foldAllSetting(){
                console.log("foldAllSetting begin");
                this.$children.forEach(v=>{
                    if(v.$attrs["data-setting"]){
                        v.$children[0].showSetting = false;
                    }
                })
                console.log("foldAllSetting end");
            },
            unfoldAllSetting(){
                this.$children.forEach(v=>{
                    if(v.$attrs["data-setting"]){
                        v.$children[0].showSetting = true;
                    }
                })
            },
            updateBaseURL() {
                ajax.setBaseURL(this.baseURL);
                this.$message.success("设置api接口地址前缀成功");
            },
            showSetting() {
                // this.$message(JSON.stringify(this.$data, null, 2))
                this.configJson = JSON.stringify(this.sForm, null, 2);
                this.dialogVisible = true;
            },
            async loadSetting() {
                let config = await Config.getConfig();
                if (config.bSuccess) {
                    Object.assign(this.sForm, config.data);
                    this.$message.success("读取服务器配置成功");
                }else{
                    this.$message.error(config.msg);
                }
            },
            async saveSetting() {
                let res = await Config.updateConfig(this.sForm);
                if(res.bSuccess) {
                    this.$message.success("保存配置成功");
                }else{
                    this.$message.error(res.msg);
                }
            },
            async reloadServer() {
                let res = await Config.reloadServer(this.sForm);
                if(res.bSuccess) {
                    this.$message.success("reload配置成功");
                }else{
                    this.$message.error(res.msg);
                }
            },
            updateSetting() {
                this.sForm = JSON.parse(this.configJson);
                console.log(this.configJson);
            },
            logChanged(newValue) {
                console.log(newValue);
            },

        },
        computed: {
            configSetting() {
                return JSON.stringify(this.sForm, null, 2)
            }
        }
    }
</script>
<style>
    .el-input-number {
        width: 100%;
    }

    .el-select {
        width: 100%;
    }
    .el-button+.el-button-group {
        margin-left: 10px;
    }
</style>
