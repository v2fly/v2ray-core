<template>
    <div>
        <h2>认证用户列表</h2>
        <el-table
                :data="sForm.users"
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
                    prop="email"
                    label="email">
                <template slot-scope="scope">
                    <el-input v-model="scope.row.email" placeholder="email" v-setting></el-input>
                </template>
            </el-table-column>
            <el-table-column
                    prop="secret"
                    label="secret">
                <el-tooltip placement="top" effect="light" slot="header">
                    <div slot="content">
                        <p>用户密钥。必须为 32 个字符，仅可包含 0 到 9 和 a 到 f 之间的字符。</p>
                        <p>使用此命令生成 MTProto 代理所需要的用户密钥：openssl rand -hex 16</p>
                    </div>
                    <label>secret</label>
                </el-tooltip>
                <template slot-scope="scope">
                    <el-input v-model="scope.row.secret" placeholder="secret" v-setting></el-input>
                </template>
            </el-table-column>
            <el-table-column
                    prop="level"
                    label="level">
                <template slot-scope="scope">
                    <el-input-number v-model.number="scope.row.level" placeholder="level" v-setting></el-input-number>
                </template>
            </el-table-column>
        </el-table>
        <p>目前只有第一个用户会生效</p>
    </div>
</template>

<script>
    export default {
        name: "InboundMTProtoSetting",
        model: {
            prop: 'setting',
            event: 'change'
        },
        data() {
            return {
                changedByForm: false,
                sForm: {
                    "users": [
                        {
                            "email": "love@v2ray.com",
                            "level": 0,
                            "secret": "b0cbcef5a486d9636472ac27f8e11a9d"
                        }
                    ]
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
                setting = this._.pick(setting, ["users"]);
                this.sForm = this._.defaults(setting, this.sForm);
                let users = this.sForm.users || [];
                this.sForm.users = [];
                users.forEach((user) => {
                    this.sForm.users.push(Object.assign({"email": "", "secret": "", "level": 0}, user));
                });
                this.$nextTick().then(() => {
                    this.formChanged();
                });
            },
            newUser() {
                this.sForm.users.push({
                    "email": "", "secret": "", "level": 0
                });
            },
            delUser(idx) {
                this.sForm.users.splice(idx, 1);
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
