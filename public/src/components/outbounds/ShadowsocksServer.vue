<template>
    <div class="vnext">
        <el-form :inline="false" label-width="120px" class="text-left">
            <el-row>
                <el-col :span="8">
                    <el-form-item label="远端地址">
                        <el-input
                                v-model="sForm.address" v-setting>
                            <el-button slot="append" icon="el-icon-delete" @click="$emit('del-server', idx)"/>
                        </el-input>
                    </el-form-item>
                </el-col>
                <el-col :span="8">
                    <el-form-item label="远端端口">
                        <el-input
                                v-model.number="sForm.port" v-setting>
                        </el-input>
                    </el-form-item>
                </el-col>
                <el-col :span="8">
                    <el-form-item label="加密方式">
                        <el-select v-model="sForm.method" v-setting>
                            <el-option value="aes-256-cfb" />
                            <el-option value="aes-128-cfb" />
                            <el-option value="chacha20" />
                            <el-option value="chacha20-ietf" />
                            <el-option value="aes-256-gcm" />
                            <el-option value="aes-128-gcm" />
                            <el-option value="chacha20-poly1305" />
                            <el-option value="none" />
                        </el-select>
                    </el-form-item>
                </el-col>
            </el-row>
            <el-row>
                <el-col :span="8">
                    <el-form-item label="email">
                        <el-input v-model="sForm.email" v-setting>
                        </el-input>
                    </el-form-item>
                </el-col>
                <el-col :span="8">
                    <el-form-item label="password">
                        <el-input v-model="sForm.password" v-setting>
                        </el-input>
                    </el-form-item>
                </el-col>
                <el-col :span="5">
                    <el-form-item label="level">
                        <el-input v-model.number="sForm.level" v-setting>
                        </el-input>
                    </el-form-item>
                </el-col>
                <el-col :span="3">
                    <el-form-item label="ota" label-width="60px">
                        <el-switch v-model="sForm.ota" v-setting></el-switch>
                    </el-form-item>
                </el-col>
            </el-row>


        </el-form>
    </div>
</template>

<script>
    import * as G from '@/consts';

    export default {
        name: "ShadowsocksServer",
        model: {
            prop: 'setting',
            event: 'change'
        },
        methods: {

            getSettings() {
                return Object.assign({}, this.sForm);
            },
            formChanged() {
                this.changedByForm = true;
                this.$emit("change", this.getSettings());
            },
            fillDefaultValue(setting) {
                setting = setting || {};
                setting = this._.pick(setting, ["id", "email", "address", "port", "users",
                    "method","password", "ota", "level"]);
                this.sForm = this._.defaults(setting, this.sForm);
                this.$nextTick().then(()=>{
                    this.formChanged();
                })
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
                    "email": "love@v2ray.com",
                    "address": "127.0.0.1",
                    "port": 1234,
                    "method": "加密方式",
                    "password": "密码",
                    "ota": false,
                    "level": 0
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
        },
    }
</script>

<style scoped>
    .vnext {
        margin-bottom: 20px;
        border-bottom: 1px solid #8c939d;
    }
</style>
