<template>
    <div>


        <el-form :inline="false" label-width="90px" class="text-left">
            <el-row :gutter="10">
                <el-col :span="8">
                    <el-form-item label="address">
                        <el-input v-model="sForm.address" v-setting/>
                    </el-form-item>
                </el-col>
                <el-col :span="8">
                    <el-form-item label="port">
                        <el-input v-model.number="sForm.port" v-setting/>
                    </el-form-item>
                </el-col>
                <el-col :span="8">
                    <el-form-item label="network">
                        <el-select v-model="sForm.network" v-setting>
                            <el-option label="tcp" value="tcp"></el-option>
                            <el-option label="udp" value="udp"></el-option>
                            <el-option label="tcp,udp" value="tcp,udp"></el-option>
                        </el-select>
                    </el-form-item>
                </el-col>
            </el-row>
            <el-row :gutter="10">
                <el-col :span="8">
                    <el-form-item label="timeout">
                        <el-input v-model.number="sForm.timeout" v-setting/>
                    </el-form-item>
                </el-col>
                <el-col :span="8">
                    <el-form-item>
                        <template v-slot:label>
                        <el-tooltip>
                            <div slot="content">当值为 true 时，dokodemo-door 会识别出由 iptables 转发而来的数据，并转发到相应的目标地址。详见 传输配置 中的 tproxy 设置。</div>
                            <label>followRedirect</label>
                        </el-tooltip>
                        </template>
                        <el-switch v-model="sForm.followRedirect" v-setting/>
                    </el-form-item>
                </el-col>
                <el-col :span="8">
                    <el-form-item label="userLevel">
                        <el-input-number v-model.number="sForm.network" v-setting/>
                    </el-form-item>
                </el-col>
            </el-row>
        </el-form>
    </div>
</template>

<script>
    export default {
        name: "VmessDokodemoDoorSetting",
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
                setting = this._.pick(setting, ["address", "port", "network", "timeout", "followRedirect", "userLevel"]);
                this.sForm = this._.defaults(setting, this.sForm);

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
                    "address": "8.8.8.8",
                    "port": 53,
                    "network": "tcp",
                    "timeout": 0,
                    "followRedirect": false,
                    "userLevel": 0
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
