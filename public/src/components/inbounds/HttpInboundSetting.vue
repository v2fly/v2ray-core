<template>
    <div>
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
                    prop="user"
                    label="user">
                <template slot-scope="scope">
                    <el-input v-model="scope.row.user" placeholder="user" v-setting></el-input>
                </template>
            </el-table-column>
            <el-table-column
                    prop="pass"
                    label="pass">
                <template slot-scope="scope">
                    <el-input v-model="scope.row.pass" placeholder="password" v-setting></el-input>
                </template>
            </el-table-column>
        </el-table>
        <el-form :inline="false" label-width="90px" class="text-left">
            <el-row>
                <el-col :span="8">
                    <el-form-item label="timeout">
                        <el-input-number
                                v-model.number="sForm.timeout" v-setting>
                        </el-input-number>
                    </el-form-item>
                </el-col>
                <el-col :span="8">
                    <el-form-item label="userLevel">
                        <el-input-number v-model.number="sForm.userLevel" v-setting/>
                    </el-form-item>
                </el-col>
                <el-col :span="8">
                    <el-form-item label="allowTransparent" label-width="130px">
                        <el-switch v-model="sForm.allowTransparent" v-setting/>
                    </el-form-item>
                </el-col>
            </el-row>
        </el-form>
    </div>
</template>

<script>
    export default {
        name: "HttpInboundSetting",
        model: {
            prop: 'setting',
            event: 'change'
        },
        data() {
            return {
                changedByForm: false,
                sForm: {
                    "timeout": 0,
                    "accounts": [],
                    "allowTransparent": false,
                    "userLevel": 0
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
            this.fillDefaultValue(this.setting);
        },
        mounted() {
            // $(this.$el).on("change", "input", ()=>{
            //     this.formChanged();
            // });

        },
        methods: {
            fillDefaultValue(setting) {
                setting = this.setting || {};
                setting = this._.pick(setting, ["timeout", "accounts", "allowTransparent", "userLevel"]);
                this.sForm = this._.defaults(setting, this.sForm);
                let accounts = this.sForm.accounts || [];
                this.sForm.accounts = [];
                accounts.forEach((account) => {
                    this.sForm.accounts.push(Object.assign({"user": "", "pass": ""}, account));
                });
                this.$nextTick().then(() => {
                    this.formChanged();
                });
            },
            newUser() {
                this.sForm.accounts.push({
                    "user": "",
                    "pass": ""
                });
            },
            delUser(idx) {
                this.sForm.accounts.splice(idx, 1);
                this.formChanged();
            },
            getSettings() {
                let setting = Object.assign({}, this.sForm);
                return setting;
            },
            formChanged() {
                this.changedByForm = true;
                this.$emit("change", this.getSettings());
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

</style>
