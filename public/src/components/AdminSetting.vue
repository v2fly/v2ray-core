<template>
    <setting-card title="Admin管理台设置" :show-enable="true" @update:enableSetting="enableSettingChanged"
                  :enableSetting.sync="enableSetting">
        <el-form ref="form" label-width="80px">
            <el-form-item label="监听端口">
                <template v-slot:label>
                    <el-tooltip>
                        <div slot="content">监听端口可以只配置端口或者指定为监听ip+端口形式</div>
                        <label>监听端口</label>
                    </el-tooltip>
                </template>
                <el-input v-model="sForm.addr" v-setting></el-input>
            </el-form-item>
            <el-form-item label="contextPath">
                <template v-slot:label>
                    <el-tooltip >
                        <div slot="content" >admin web前缀路径</div>
                        <label>contextPath</label>
                    </el-tooltip>
                </template>
                <el-input v-model="sForm.contextPath" v-setting></el-input>
            </el-form-item>
            <el-form-item label="publicPath">
                <el-input v-model="sForm.publicPath" placeholder="应用静态资源文件路径" v-setting></el-input>
            </el-form-item>
        </el-form>
        <el-table
                :data="sForm.accounts"
                style="width: 100%;margin-bottom: 10px;">
            <el-table-column
                    label="op"
                    width="60">
                <template slot="header">
                    <i type="primary" class="el-icon-plus" @click="newUser"></i>
                </template>
                <template slot-scope="scope">
                    <i class="el-icon-delete" @click="delUser(scope.$index)"></i>
                </template>
            </el-table-column>
            <el-table-column
                    prop="userName"
                    label="userName">
                <template slot-scope="scope">
                    <el-input v-model="scope.row.userName" placeholder="userName" v-setting></el-input>
                </template>
            </el-table-column>
            <el-table-column
                    prop="password"
                    label="password">
                <template slot-scope="scope">
                    <el-input v-model="scope.row.password" placeholder="password" v-setting></el-input>
                </template>
            </el-table-column>
        </el-table>
    </setting-card>

</template>

<script>

    export default {
        name: "AdminSetting",
        model:{
            prop:"setting",
            event:"change"
        },
        data() {
            return {
                "enableSetting": true,
                changedByForm: false,
                sForm: {
                    "addr":":8035",
                    "contextPath": "/v2ray",
                    "publicPath": "../public/dist",
                    "accounts": [],
                }
            }
        },
        created() {
            this.fillDefaultValue(this.setting);
        },
        mounted() {

        },
        methods: {
            newUser() {
                this.sForm.accounts.push({
                    "userName": "",
                    "password": ""
                });
            },
            delUser(idx) {
                this.sForm.accounts.splice(idx, 1);
            },
            enableSettingChanged() {
                this.$nextTick(() => {
                    this.formChanged();
                });
            },
            fillDefaultValue(setting){
                setting = setting || {};
                Object.assign(this.sForm, setting);
                this.$nextTick().then(()=>{
                    this.formChanged();
                });
            },
            formChanged() {
                let setting = this.getSettings();
                if(this.setting !== setting) {
                    this.changedByForm = true;
                    this.$emit("change", setting);
                }

            },
            getSettings() {
                return Object.assign({},this.sForm);
            },
        },
        watch: {
            setting: {
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
        props: {
            setting: {
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
