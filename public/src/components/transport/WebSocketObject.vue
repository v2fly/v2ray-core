<template>
    <setting-card title="WebSocket传输数据" :enable-setting.sync="enableSetting"
                  @update:enableSetting="enableSettingChanged">
        <el-form :inline="true" label-width="90px" class="text-left">
            <el-form-item label="path">
                <el-input v-model="sForm.path" v-setting/>
            </el-form-item>
            <el-form-item label="acceptProxyProtocol" label-width="140px">
                <el-switch v-model="sForm.acceptProxyProtocol" v-setting/>
            </el-form-item>
        </el-form>
        <h2>headers</h2>
        <el-table
                :data="headers"
                style="width: 100%;margin-bottom: 10px;">
            <el-table-column
                    label="op"
                    width="60">
                <template slot="header">
                    <i type="primary" class="el-icon-plus" @click="newHeader"></i>
                </template>
                <template slot-scope="scope">
                    <i class="el-icon-delete" @click="delHeader(scope.$index)"></i>
                </template>
            </el-table-column>
            <el-table-column
                    prop="key"
                    label="key">
                <template slot-scope="scope">
                    <el-input v-model="scope.row.key" placeholder="key" v-setting></el-input>
                </template>
            </el-table-column>
            <el-table-column
                    prop="value"
                    label="value">
                <template slot-scope="scope">
                    <el-input v-model="scope.row.value" placeholder="value" v-setting></el-input>
                </template>
            </el-table-column>
        </el-table>

    </setting-card>
</template>

<script>
    export default {
        name: "WebSocketObject",
        model: {
            prop: 'setting',
            event: 'change'
        },
        data() {
            return {
                enableSetting: false,
                changedByForm: false,
                headers: [],
                sForm: {
                    "acceptProxyProtocol": false,
                    "path": "/",
                    "headers": {}
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
                let headers = [];
                for (let k in this.sForm.headers) {
                    headers.push({key: k, value: this.sForm.headers[k]});
                }
                this.headers = headers;
                this.formChanged();
            },
            newHeader() {
                this.headers.push({
                    "key": "",
                    "value": ""
                });
            },
            delHeader(idx) {
                this.headers.splice(idx, 1);
                this.formChanged();
            },
            getSettings() {
                if (!this.enableSetting) {
                    return null;
                }
                let headersMap = {};
                this.headers.forEach(h => {
                    headersMap[h.key] = h.value;
                });
                this.sForm.headers = headersMap;
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
