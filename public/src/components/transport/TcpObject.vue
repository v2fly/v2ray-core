<template>
    <setting-card title="TCP传输数据" :enable-setting.sync="enableSetting" @update:enableSetting="enableSettingChanged">
        <el-form :inline="true" label-width="90px" class="text-left">
            <el-form-item label="acceptProxyProtocol" label-width="140px">
                <el-switch v-model="sForm.acceptProxyProtocol" v-setting/>
            </el-form-item>
            <el-form-item label="headerType" label-width="100px">
                <el-select v-model="headerType" placeholder="header type">
                    <el-option value="none"/>
                    <el-option value="http"/>
                </el-select>
            </el-form-item>
        </el-form>
        <div v-if="headerType=='http'">
            <HttpRequestObject v-model="sForm.header.request" v-setting/>
            <HttpResponseObject v-model="sForm.header.response" v-setting/>
        </div>

    </setting-card>
</template>

<script>
    import HttpRequestObject from "@/components/transport/HttpRequestObject";
    import HttpResponseObject from "@/components/transport/HttpResponseObject";
    export default {
        name: "TcpObject",
        components:{HttpRequestObject, HttpResponseObject},
        model: {
            prop: 'setting',
            event: 'change'
        },
        data() {
            return {
                enableSetting: false,
                changedByForm: false,
                headerType: "none",
                sForm: {
                    "acceptProxyProtocol": false,
                    "header":{
                        "type":"none",
                        "request":{},
                        "response":{}
                    }
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
