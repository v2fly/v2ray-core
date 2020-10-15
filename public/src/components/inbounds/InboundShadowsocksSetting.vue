<template>
    <div>
        <el-form :inline="false" label-width="120px" class="text-left">
            <el-row>
                <el-col :span="8">
                    <el-form-item label="email">
                        <el-input v-model="sForm.email" v-setting>
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
                <el-col :span="8">
                    <el-form-item label="password">
                        <el-input v-model="sForm.password" v-setting>
                        </el-input>
                    </el-form-item>
                </el-col>
            </el-row>
            <el-row>

                <el-col :span="8">
                    <el-form-item label="level">
                        <el-input v-model.number="sForm.level" v-setting>
                        </el-input>
                    </el-form-item>
                </el-col>
                <el-col :span="8">
                    <el-form-item label="network">
                        <el-select v-model="sForm.network" v-setting>
                            <el-option value="tcp" />
                            <el-option value="udp" />
                            <el-option value="tcp,udp" />
                        </el-select>
                    </el-form-item>
                </el-col>
                <el-col :span="8">
                    <el-form-item label="ota" >
                        <el-switch v-model="sForm.ota" v-setting></el-switch>
                    </el-form-item>
                </el-col>
            </el-row>


        </el-form>
    </div>
</template>

<script>
    export default {
        name: "InboundShadowsocksSetting",
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
                setting = this.setting || {};
                setting = this._.pick(setting, ["email", "method", "password", "level", "ota", "network"]);
                this.sForm = this._.defaults(setting, this.sForm);
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
                    "email": "love@v2ray.com",
                    "method": "aes-128-cfb",
                    "password": "",
                    "level": 0,
                    "ota": true,
                    "network": "tcp"
                }
            }
        },
        props: {
            setting: {
                type: Object
            }
        },
        computed: {
        },
    }
</script>

<style scoped>

</style>
