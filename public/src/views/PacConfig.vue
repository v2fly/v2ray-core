<template>
    <div>
        <el-form :inline="false" label-position="right" label-width="120px">
            <el-form-item label="PAC地址">
                <label>{{pacUrl}}</label>
            </el-form-item>
            <el-form-item label="proxy地址">
                <el-input title="代理地址，格式为: PROXY localhost:10809;SOCK5 localhost:10808;SOCKS 127.0.0.1:1080; DIRECT;"
                          placeholder="代理地址，格式为: PROXY localhost:10809;SOCK5 localhost:10808;SOCKS 127.0.0.1:1080; DIRECT;"
                          v-model="config.proxy"/>
            </el-form-item>
            <el-form-item label="下载代理">
                <el-input placeholder="下载gfwlist文件，geosite,geoip文件时使用的http代理地址, http://localhost:20809"
                          v-model="config.gfwProxy"/>
            </el-form-item>
            <el-form-item label="自定义规则">
                <el-input type="textarea" rows="8"
                          v-model="config.userRule"/>
            </el-form-item>
            <el-form-item>
                <el-button @click="downloadGfwList">更新GFWList</el-button>
                <el-button @click="downloadGeoDat">更新geo文件</el-button>
                <el-button @click="savePac">保存PAC配置</el-button>
                <el-button @click="viewPacContent">查看PAC文件</el-button>
            </el-form-item>
        </el-form>

        <el-dialog :visible.sync="dialogVisible" title="PAC内容" center>
            <el-input type="textarea" placeholder="placeholder" rows="10" v-model="pacContent"></el-input>
            <div slot="footer" class="dialog-footer">
                <el-button @click="dialogVisible = false">关闭</el-button>
            </div>
        </el-dialog>

    </div>
</template>

<script>
    import Pac from "@/api/Pac";
    import ajax from "@/api/lib/ajax"

    export default {
        name: "TrafficMonitor",
        data() {
            return {
                pacUrl: ajax.getBaseURL() + "/api/pac",
                pacContent: "",
                dialogVisible: false,
                config: {
                    proxy: "",
                    userRule: "",
                    gfwProxy: "",
                }
            }
        },
        created() {
            this.loadConfig();
        },
        destroyed() {

        },
        methods: {
            async loadConfig() {
                let res = await Pac.loadConfig();
                if (res.bSuccess) {
                    Object.assign(this.config, res.data);
                } else {
                    this.$message.error(res.msg);
                }
            },
            async downloadGfwList() {
                let res = await Pac.downloadGfwList(this.config);
                if (res.bSuccess) {
                    this.$message.success("更新gfwlist成功");
                } else {
                    this.$message.error(res.msg);
                }
            },
            async downloadGeoDat() {
                let res = await Pac.downloadGeoDat(this.config);
                if (res.bSuccess) {
                    this.$message.success("更新geosite.dat,geoip.dat成功");
                } else {
                    this.$message.error(res.msg);
                }
            },
            async savePac() {
                let res = await Pac.savePac(this.config);
                if (res.bSuccess) {
                    this.$message.success("保存pac配置成功");
                } else {
                    this.$message.error(res.msg);
                }
            },
            async viewPacContent() {
                let res = await Pac.getPacContent();
                if (res.bSuccess) {
                    this.pacContent = res.data;
                    this.dialogVisible = true;
                } else {
                    this.$message.error(res.msg);
                }

            }

        },
    }
</script>

<style scoped>

</style>
