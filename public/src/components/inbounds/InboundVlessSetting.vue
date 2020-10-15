<template>
    <div>
        <h2>用户列表:</h2>
        <el-table
                :data="sForm.clients"
                style="width: 100%">
            <el-table-column
                    prop="id"
                    label="op"
                    width="40">
                <template slot="header">
                    <i type="primary" class="el-icon-plus" @click="newUser"></i>
                </template>
                <template slot-scope="scope">
                    <i class="el-icon-delete" @click="delUser(scope.$index)"></i>
                </template>
            </el-table-column>
            <el-table-column
                    prop="id"
                    label="id"
                    width="380">
                <template slot-scope="scope">
                    <el-input v-model="scope.row.id" placeholder="id" v-setting>
                        <el-button icon="el-icon-refresh" slot="append" @click="scope.row.id=newUUID()"></el-button>
                    </el-input>
                </template>
            </el-table-column>
            <el-table-column
                    prop="level"
                    label="level"
                    width="200">
                <template slot-scope="scope">
                    <el-input-number v-model.number="scope.row.level" placeholder="level" v-setting></el-input-number>
                </template>
            </el-table-column>
            <el-table-column
                    prop="email"
                    label="email">
                <template slot-scope="scope">
                    <el-input v-model="scope.row.email" placeholder="email" v-setting></el-input>
                </template>
            </el-table-column>
        </el-table>
        <h2>回落分流配置:</h2>
        <el-table
                :data="sForm.fallbacks"
                style="width: 100%">
            <el-table-column
                    prop="id"
                    label="op"
                    width="40">
                <template slot="header">
                    <i type="primary" class="el-icon-plus" @click="newFallback"></i>
                </template>
                <template slot-scope="scope">
                    <i class="el-icon-delete" @click="delFallback(scope.$index)"></i>
                </template>
            </el-table-column>
            <el-table-column
                    prop="alpn"
                    label="alpn"
                    width="200">
                <template slot-scope="scope">
                    <el-select clearable v-model="scope.row.alpn" placeholder="alpn" allow-create filterable v-setting>
                        <el-option value="http/1.1"/>
                    </el-select>
                </template>
            </el-table-column>
            <el-table-column
                    prop="dest"
                    label="dest"
                    width="200">
                <template slot-scope="scope">
                    <el-input v-model="scope.row.dest" placeholder="dest" v-setting></el-input>
                </template>
            </el-table-column>
            <el-table-column
                    prop="xver"
                    label="xver"
                    width="200">
                <template slot-scope="scope">
                    <el-select v-model="scope.row.xver" placeholder="xver" v-setting>
                        <el-option value="0" />
                        <el-option value="1" />
                        <el-option value="2" />
                    </el-select>
                </template>
            </el-table-column>
            <el-table-column
                    prop="path"
                    label="path">
                <template slot-scope="scope">
                    <el-input v-model="scope.row.path" placeholder="path" v-setting></el-input>
                </template>
            </el-table-column>
        </el-table>
    </div>
</template>

<script>
    import {v4 as uuidv4} from 'uuid';

    export default {
        name: "VlessInboundSetting",
        model: {
            prop: 'setting',
            event: 'change'
        },
        methods: {
            newUUID() {
                return uuidv4();
            },
            newFallback() {
                this.sForm.fallbacks.push({
                    "dest": "80", "path": "", "alpn": "", "xver": 0
                });
            },
            delFallback(idx) {
                this.sForm.fallbacks.splice(idx, 1);
            },
            newUser() {
                this.sForm.clients.push({
                    "id": uuidv4(),
                    "level": 0,
                    "email": ""
                });
            },
            delUser(idx) {
                this.sForm.clients.splice(idx, 1);
            },
            getSettings() {
                return Object.assign({}, this.sForm);
            },
            formChanged() {
                this.changedByForm = true;
                this.$emit("change", this.getSettings());
            },
            fillDefaultValue(setting) {
                setting = this.setting || {};
                setting = this._.pick(setting, ["clients", "decryption", "fallbacks"]);
                this.sForm = this._.defaults(setting, this.sForm);

                let clients = this.sForm.clients || [];
                this.sForm.clients = [];
                clients.forEach((client) => {
                    this.sForm.clients.push(Object.assign({"level": 0, "email": "", "id": ""}, client));
                });

                let fallbacks = this.sForm.fallbacks || [];
                this.sForm.fallbacks = [];
                fallbacks.forEach((f) => {
                    this.sForm.fallbacks.push(Object.assign({"dest": "80", "path": "", "alpn": "", "xver": 0}, f));
                });
                this.$nextTick().then(()=>{
                    this.formChanged();
                });
            }
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

        created() {
            this.fillDefaultValue(this.setting);
        },
        mounted() {
            // $(this.$el).on("change", "input", ()=>{
            //     this.formChanged();
            // });
        },

        data() {
            return {
                changedByForm: false,
                sForm: {
                    "clients": [],
                    "decryption": "none",
                    "fallbacks": []
                }
            }
        },
        props: {
            setting: {
                type: Object
            }
        },
        computed: {
            clientDefault: {
                get() {
                    return this.sForm.default;
                },
                set(newDef) {
                    this.sForm.default = newDef;
                }
            }
        },
    }
</script>

<style scoped>

</style>
