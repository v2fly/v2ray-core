<template>
    <setting-card title="API接口开放设置" :show-enable="true"
                  @update:enableSetting="enableSettingChanged"
                  :enableSetting.sync="enableSetting">
        <el-form ref="form" :model="sForm" label-width="80px">
            <el-form-item label="tag">
                <el-input v-model="sForm.tag" v-setting></el-input>
            </el-form-item>
            <el-form-item label="服务">
                <el-checkbox-group v-model="sForm.services" align="left">
                    <el-checkbox label="HandlerService" value="HandlerService" v-setting name="services"></el-checkbox>
                    <el-checkbox label="LoggerService" value="LoggerService" v-setting name="services"></el-checkbox>
                    <el-checkbox label="StatsService" value="StatsService" v-setting name="services"></el-checkbox>
                </el-checkbox-group>
            </el-form-item>
        </el-form>
    </setting-card>
</template>

<script>
    import SettingCard from "@/components/SettingCard";

    export default {
        name: "ApiSetting",
        components: {SettingCard},
        model: {
            prop: 'api',
            event: 'change'
        },
        created() {
            const setting = this.api || {};
            Object.assign(this.sForm, setting);
            this.formChanged();
        },
        mounted() {


        },
        watch: {
            api(val) {
                if(this.changedByForm){
                    this.changedByForm = false;
                    return;
                }
                const setting = val || {};
                Object.assign(this.sForm, setting);
                this.formChanged();
            }
        },
        methods: {
            getSettings() {
                if (!this.enableSetting) {
                    return null;
                }
                return Object.assign({}, this.sForm);
            },
            enableSettingChanged() {
                this.$nextTick(() => {
                    this.formChanged();
                });

            },
            formChanged() {

                if (this.enableSetting) {
                    this.$store.commit("setApiTag", this.sForm.tag);
                } else {
                    this.$store.commit("setApiTag", null);
                }

                this.changedByForm = true;
                this.$emit("change", this.getSettings());
            }
        },
        data() {
            return {
                sForm: {
                    "tag": "api",
                    "services": [],
                },
                changedByForm: false,
                "enableSetting": true,
            }
        },
        props: {
            api: {
                type: Object
            }
        }
    }
</script>

<style scoped>
    .el-select {
        width: 100%;
    }
</style>
