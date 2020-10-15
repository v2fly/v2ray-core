<template>
    <div>

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
                    prop="key" width="230"
                    label="key">
                <template slot-scope="scope">
                    <el-select v-model="scope.row.key" filterable allow-create placeholder="key" v-setting>
                        <el-option v-for="h in headerKeys" :value="h" :key="h"/>
                    </el-select>
                </template>
            </el-table-column>
            <el-table-column
                    prop="value"
                    label="value">
                <template slot-scope="scope">
                    <div v-for="(v,idx) in scope.row.value" :key="idx" class="el-input el-input-group el-input-group--append">
                        <input :value="v" @change="scope.row.value[idx] = $event.currentTarget.value" v-setting
                               type="text" autocomplete="off" placeholder="value" class="el-input__inner">
                        <div class="el-input-group__append">
                            <el-button icon="el-icon-plus" @click="scope.row.value.push('')"
                                       v-if="idx==0"></el-button>
                            <el-button icon="el-icon-delete" @click="scope.row.value.splice(idx,1)"
                                       v-if="idx>0"></el-button>
                        </div>
                    </div>
                </template>
            </el-table-column>
        </el-table>

    </div>
</template>

<script>
    import * as G from '@/consts'
    export default {
        name: "HttpHeaders",
        model: {
            prop: 'setting',
            event: 'change'
        },
        data() {
            return {
                enableSetting: true,
                changedByForm: false,
                headers: [],
                headerKeys: G.HTTP_HEADER_KEYS,
                sForm: {

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
                this.sForm = Object.assign({}, setting);
                let headers = [];
                for (let k in this.sForm) {
                    let v = this.sForm[k];
                    if (!Array.isArray(v)) {
                        v = [v];
                    }
                    headers.push({key: k, value: v});
                }
                this.headers = headers;
                this.formChanged();
            },
            newHeader() {
                this.headers.push({
                    "key": "",
                    "value": [""]
                });
            },
            delHeader(idx) {
                this.headers.splice(idx, 1);
                this.formChanged();
            },
            headerValueChange(event) {
                console.log(event);
            },
            getSettings() {
                if (!this.enableSetting) {
                    return null;
                }
                let headersMap = {};
                this.headers.forEach(h => {
                    headersMap[h.key] = h.value;
                });
                this.sForm = headersMap;
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
