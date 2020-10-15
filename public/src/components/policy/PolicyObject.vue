<template>
    <setting-card title="Policy 本地策略" :show-enable="true"
                  @update:enableSetting="enableSettingChanged"
                  :enableSetting.sync="enableSetting">
        <template v-slot:header-buttons>
            <i class="el-icon-plus" style="margin-right:10px;" @click="newLevel"></i>
        </template>
        <el-form label-width="150px">
            <el-row >
                <el-col :span="6">
                    <el-form-item label="statsInboundUplink" >
                        <el-tooltip slot="label" effect="light" placement="top">
                            <div slot="content">
                                当值为 true 时，开启所有入站代理的上行流量统计。
                            </div>
                            <label>statsInboundUplink</label>
                        </el-tooltip>
                        <el-switch  v-model="sForm.system.statsInboundUplink" v-setting />
                    </el-form-item>
                </el-col>
                <el-col :span="6">
                    <el-form-item label="statsInboundDownlink" >
                        <el-tooltip slot="label" effect="light" placement="top">
                            <div slot="content">
                                当值为 true 时，开启所有入站代理的下行流量统计。
                            </div>
                            <label>statsInboundDownlink</label>
                        </el-tooltip>
                        <el-switch  v-model="sForm.system.statsInboundDownlink" v-setting />
                    </el-form-item>
                </el-col>
                <el-col :span="6">
                    <el-form-item label="statsOutboundUplink" >
                        <el-tooltip slot="label" effect="light" placement="top">
                            <div slot="content">
                                （ V2Ray 4.26.0+ ）当值为 true 时，开启所有出站代理的上行流量统计。
                            </div>
                            <label>statsOutboundUplink</label>
                        </el-tooltip>
                        <el-switch  v-model="sForm.system.statsOutboundUplink" v-setting />
                    </el-form-item>
                </el-col>
                <el-col :span="6">
                    <el-form-item label="statsOutboundDownlink" >
                        <el-tooltip slot="label" effect="light" placement="top">
                            <div slot="content">
                                （ V2Ray 4.26.0+ ） 当值为 true 时，开启所有出站代理的下行流量统计。
                            </div>
                            <label>statsOutboundDownlink</label>
                        </el-tooltip>
                        <el-switch  v-model="sForm.system.statsOutboundDownlink" v-setting />
                    </el-form-item>
                </el-col>
            </el-row>
        </el-form>
        <LevelPolicyObject v-setting v-for="(l,idx) in levelKeys" :key="l" :level="l" :setting="sForm.levels[l]"
                           @del-level="delLevel(l, idx)"
                           @change="sForm.levels[l]=$event" />

    </setting-card>
</template>

<script>

    import LevelPolicyObject from "@/components/policy/LevelPolicyObject";
    export default {
        name: "PolicyObject",
        components: {LevelPolicyObject},
        model: {
            prop: 'setting',
            event: 'change'
        },
        created() {
            this.fillDefaultValue(this.setting);
        },
        mounted() {
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
        methods: {
            delLevel(level, idx){
                this.levelKeys.splice(idx, 1);
                delete this.sForm.levels[level];
                this.formChanged();
            },
            newLevel() {
                this.$prompt('请输入用户level值', '提示', {
                    confirmButtonText: '确定',
                    cancelButtonText: '取消',
                    inputPattern: /^\d+$/,
                    inputErrorMessage: '用户level值不正确'
                }).then(({ value }) => {
                    if(this.sForm.levels[value]){
                        this.$message({
                            type: 'error',
                            message: '用户level='+value+'定义已存在'
                        });
                        return;
                    }
                    this.levelKeys.push(value);
                    this.$set(this.sForm.levels, value, {});
                });
            },
            fillDefaultValue(setting) {
                setting = setting || {};
                if (setting.system || setting.levelKeys) {
                    this.enableSetting = true;
                } else {
                    this.enableSetting = false;
                }
                this.levelKeys = [];
                Object.assign(this.sForm, setting);
                for(let l in this.sForm.levels) {
                    this.levelKeys.push(l);
                }
                this.$nextTick(() => {
                    this.formChanged();
                });
            },
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
                let setting = this.getSettings();
                if(setting == null && this.setting == null) {
                    return;
                }
                this.changedByForm = true;
                this.$emit("change", setting);
            }
        },
        data() {
            return {
                levelKeys:[],
                sForm: {
                    "levels": {
                    },
                    "system": {
                        "statsInboundUplink": false,
                        "statsInboundDownlink": false,
                        "statsOutboundUplink": false,
                        "statsOutboundDownlink": false
                    }
                },
                changedByForm: false,
                "enableSetting": false,
            }
        },
        props: {
            setting: {
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
