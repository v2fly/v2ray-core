<template>
    <setting-card title="HTTP/2传输数据" :enable-setting.sync="enableSetting" @update:enableSetting="enableSettingChanged">
        <el-form :inline="false" label-width="90px" class="text-left">
            <el-form-item label="path">
                <el-input v-model="sForm.path" v-setting/>
            </el-form-item>
            <el-form-item label="host">
                <el-input type="textarea" placeholder="host" rows="3" v-model="hostValue" v-setting />
            </el-form-item>
        </el-form>

    </setting-card>
</template>

<script>
    export default {
        name: "Http2Object",
        model: {
            prop: 'setting',
            event: 'change'
        },
        data() {
            return {
                enableSetting: false,
                changedByForm: false,
                sForm: {
                    "path": "/",
                    "host": []
                }

            }
        },
        computed:{
            hostValue:{
                set(val){
                    this.sForm.host = val.split("\n");
                },
                get() {
                    return this.sForm.host.join("\n");
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
            let setting = this.setting || {};
            this.fillDefaultValue(setting);

        },
        mounted() {
        },
        methods: {
            fillDefaultValue(setting){
                setting = this.setting || {};
                if (this._.isEmpty(setting)) {
                    this.enableSetting = false;
                } else {
                    this.enableSetting = true;
                }
                Object.assign(this.sForm, setting);
                this.formChanged();
            },

            getSettings() {
                if (!this.enableSetting) {
                    return null;
                }
                let setting = Object.assign({}, this.sForm);
                return setting;
            },
            formChanged() {
                this.changedByForm = true;
                this.$emit("change", this.getSettings());
            },
            enableSettingChanged() {
                this.$nextTick(() => {
                    this.formChanged();
                });

            },
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
