<template>
    <div>

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
                    width="140">
                <template slot-scope="scope">
                    <el-input-number v-model="scope.row.level" placeholder="level" v-setting></el-input-number>
                </template>
            </el-table-column>
            <el-table-column
                    prop="alterId"
                    label="alterId"
                    width="140">
                <template slot-scope="scope">
                    <el-input-number v-model="scope.row.alterId" placeholder="alterId" v-setting></el-input-number>
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
        <el-form :inline="false" label-width="90px" class="text-left" style="margin-top:20px;">
            <el-row>
                <el-col :span="8">
                    <el-form-item label="detour">
                        <el-input
                                v-model="sForm.detour.to" v-setting>
                        </el-input>
                    </el-form-item>
                </el-col>
                <el-col :span="8">
                    <el-form-item label="level">
                        <el-input-number v-model="clientDefault.level" v-setting/>
                    </el-form-item>
                </el-col>
                <el-col :span="8">
                    <el-form-item label="alterId">
                        <el-input-number v-model="clientDefault.alterId" v-setting/>
                    </el-form-item>
                </el-col>
                <el-col :span="8">
                    <el-form-item label="禁止不安全加密" label-width="120px">
                        <el-switch
                                v-model="sForm.disableInsecureEncryption" v-setting>
                        </el-switch>
                    </el-form-item>
                </el-col>
            </el-row>
        </el-form>
    </div>
</template>

<script>
    import {v4 as uuidv4 } from 'uuid';
    export default {
        name: "VmessInboundSetting",
        model: {
            prop: 'setting',
            event: 'change'
        },
        methods: {
            newUUID() {
                this.$nextTick().then(()=>{
                    this.formChanged();
                });
                return uuidv4();
            },
            newUser() {
                this.sForm.clients.push({
                    "id": uuidv4(),
                    "level": 0,
                    "alterId": 4,
                    "email": ""
                });
            },
            delUser(idx) {
                this.sForm.clients.splice(idx, 1);
                this.formChanged();
            },
            getSettings() {
                let setting = this._.cloneDeep(this.sForm);
                if(setting.detour.to==""){
                    delete setting.detour;
                    delete setting.default;
                }
                return setting;
            },
            formChanged() {
                this.changedByForm = true;
                this.$emit("change", this.getSettings());
            },
            fillDefaultValue(setting) {
                setting = this.setting || {};
                let defDefault = this.sForm.default;
                let defDetour = this.sForm.detour;
                setting = this._.pick(setting, ["clients", "disableInsecureEncryption", "default", "detour"]);
                this.sForm = this._.defaults(setting, this.sForm);
                this._.defaults(this.sForm.default, defDefault);
                this._.defaults(this.sForm.detour, defDetour);

                let clients = this.sForm.clients || [];
                this.sForm.clients = [];
                clients.forEach((client) => {
                    this.sForm.clients.push(Object.assign({"level": 0, "alterId": 4, "email": "", "id": ""}, client));
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
                    "default": {
                        "level": 0,
                        "alterId": 4
                    },
                    "detour": {
                        "to": ""
                    },
                    "disableInsecureEncryption": false
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
