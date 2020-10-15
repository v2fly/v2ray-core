<template>
    <setting-card title="日志设置" :show-enable="false">
        <el-form ref="form" label-width="80px">
            <el-form-item label="访问日志">
                <template v-slot:label>
                    <el-tooltip>
                        <div slot="content">访问日志的文件地址，其值是一个合法的文件地址，如"/tmp/v2ray/_access.log"（Linux）或者"C:\\Temp\\v2ray\\_access.log"（Windows）。
                            <br/>当此项不指定或为空值时，表示将日志输出至 stdout。V2Ray 4.20 加入了特殊值none，即关闭 access log。</div>
                        <label>访问日志</label>
                    </el-tooltip>
                </template>
                <el-input v-model="sForm.access" v-setting></el-input>
            </el-form-item>
            <el-form-item label="错误日志">
                <template v-slot:label>
                    <el-tooltip >
                        <div slot="content" >错误日志的文件地址，其值是一个合法的文件地址，如"/tmp/v2ray/_error.log"（Linux）或者"C:\\Temp\\v2ray\\_error.log"（Windows）。
                            <br/>当此项不指定或为空值时，表示将日志输出至 stdout。V2Ray 4.20 加入了特殊值none，即关闭 error log（跟loglevel: "none"等价）。</div>
                        <label>错误日志</label>
                    </el-tooltip>
                </template>
                <el-input v-model="sForm.error" v-setting></el-input>
            </el-form-item>
            <el-form-item label="日志级别">
                <el-select v-model="sForm.loglevel" placeholder="请选择日志级别" v-setting>
                    <el-option v-for="level of levels" :label="level" :value="level" :key="level"></el-option>
                </el-select>
            </el-form-item>
        </el-form>
    </setting-card>

</template>

<script>
    import SettingCard from "@/components/SettingCard";
    import * as G from '@/consts'

    export default {
        name: "LogSetting",
        components: {SettingCard},
        model:{
            prop:"log",
            event:"change"
        },
        data() {
            return {
                changedByForm: false,
                levels: G.LOG_LEVELS,
                sForm: {
                    "access": "",
                    "error": "",
                    "loglevel": "warning"
                }
            }
        },
        created() {
            const settings = this.log || {};
            Object.assign(this.sForm, settings);
        },
        mounted() {

        },
        methods: {
            formChanged() {
                this.changedByForm = true;
                this.$emit("change", this.getSettings());
            },
            getSettings() {
                return Object.assign({},this.sForm);
            },
        },
        watch: {
            log: {
                handler: function (val) {
                    if(this.changedByForm){
                        this.changedByForm = false;
                        return;
                    }
                    Object.assign(this.sForm, val);
                },
                deep: false
            }
        },
        props: {
            log: {
                type: Object,
            }
        }
    }
</script>

<style scoped>
    .el-select {
        width: 100%;
    }
</style>
