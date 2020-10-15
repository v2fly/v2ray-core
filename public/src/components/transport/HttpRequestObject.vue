<template>
    <setting-card title="Http请求头部设置" :show-enable="false" :enable-setting.sync="enableSetting">
        <el-form :inline="false" label-width="90px" class="text-left">
            <div style="display:table;width:100%;">
                <div style="display:table-cell;width:200px;">
                    <el-form-item label="version">
                        <el-input v-model="sForm.version" v-setting/>
                    </el-form-item>
                </div>
                <div style="display:table-cell;width:200px;">
                    <el-form-item label="method">
                        <el-select v-model="sForm.method" filterable allow-create v-setting>
                            <el-option value="GET"/>
                            <el-option value="POST"/>
                            <el-option value="PUT"/>
                        </el-select>
                    </el-form-item>
                </div>
                <div style="display:table-cell; width:auto;">
                    <el-form-item label="path">
                        <el-select :multiple="true" v-model="sForm.path" filterable allow-create v-setting>
                            <el-option value="/"/>
                        </el-select>
                    </el-form-item>
                </div>
            </div>
        </el-form>
        <HttpHeaders v-model="sForm.headers" v-setting/>

    </setting-card>
</template>

<script>
    import HttpHeaders from "@/components/transport/HttpHeaders";
    export default {
        name: "HttpRequestObject",
        components:{HttpHeaders,},
        model: {
            prop: 'setting',
            event: 'change'
        },
        data() {
            return {
                enableSetting: true,
                changedByForm: false,
                sForm: {
                    "version": "1.1",
                    "method": "GET",
                    "path": [
                        "/"
                    ],
                    "headers":{}
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
            fillDefaultValue(setting) {
                setting = this.setting || {};
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
