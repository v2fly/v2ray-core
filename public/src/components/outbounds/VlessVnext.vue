<template>
    <div class="vnext">
        <el-form :inline="true" label-width="120px" class="text-left">
            <el-form-item>
                <el-button type="danger" size="small" icon="el-icon-delete" @click="$emit('del-vnext', idx)">删除</el-button>
            </el-form-item>
            <el-form-item label="远端地址">
                <el-input
                        v-model="sForm.address" v-setting>
                </el-input>
            </el-form-item>
            <el-form-item label="远端端口">
                <el-input
                        v-model.number="sForm.port" v-setting>
                </el-input>
            </el-form-item>
        </el-form>
        <el-table
                :data="sForm.users"
                style="width: 100%">
            <el-table-column
                    prop="id"
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
                    prop="id"
                    label="id"
                    width="320">
                <template slot-scope="scope">
                    <el-input v-model="scope.row.id" placeholder="id" v-setting></el-input>
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
                    prop="encryption"
                    label="encryption">
                <template slot-scope="scope">
                    <el-select placeholder="encryption" v-model="scope.row.encryption" v-setting>
                        <el-option value="none"/>
                    </el-select>
                </template>
            </el-table-column>
        </el-table>
    </div>
</template>

<script>
    import * as G from '@/consts';

    export default {
        name: "VlessVnext",
        model: {
            prop: 'setting',
            event: 'change'
        },
        methods: {
            newUser() {
                this.sForm.users.push({
                    "id": "",
                    "level": 0,
                    "encryption": "none"
                });
            },
            delUser(idx) {
                this.sForm.users.splice(idx, 1);
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

                setting = this._.pick(setting, ["id", "address", "port", "users"]);
                this.sForm = this._.defaults(setting, this.sForm);

                let users = this.sForm.users || [];
                this.sForm.users = [];
                users.forEach((user) => {
                    this.sForm.users.push(Object.assign({
                        "id": "",
                        "level": 0,
                        "encryption": "none"
                    }, user));
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
                securities: G.SECURITIES,
                changedByForm: false,
                sForm: {
                    "address": "127.0.0.1",
                    "port": 37192,
                    "users": [{
                        "id": "",
                        "level": 0,
                        "encryption": "none"
                    }]
                }
            }
        },
        props: {
            setting: {
                type: Object
            },
            idx: {
                type: Number,
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
    .vnext {
        margin-bottom: 20px;
    }
</style>
