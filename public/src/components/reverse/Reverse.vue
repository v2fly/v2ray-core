<template>
    <setting-card title="Reverse反向代理" :show-enable="true"
                  @update:enableSetting="enableSettingChanged"
                  :enableSetting.sync="enableSetting">
        <h2>bridge</h2>
        <el-table
                :data="sForm.bridges"
                style="width: 100%;margin-bottom: 10px;">
            <el-table-column
                    label="op"
                    width="60">
                <template slot="header">
                    <i type="primary" class="el-icon-plus" @click="newBridge"></i>
                </template>
                <template slot-scope="scope">
                    <i class="el-icon-delete" @click="delBridge(scope.$index)"></i>
                </template>
            </el-table-column>
            <el-table-column
                    prop="tag"
                    label="tag">
                <template slot-scope="scope">
                    <el-input v-model="scope.row.tag" placeholder="tag" v-setting></el-input>
                </template>
            </el-table-column>
            <el-table-column
                    prop="domain"
                    label="domain">
                <template slot-scope="scope">
                    <el-input v-model="scope.row.domain" placeholder="domain" v-setting></el-input>
                </template>
            </el-table-column>
        </el-table>

        <h2>portals:</h2>
        <el-table
                :data="sForm.portals"
                style="width: 100%;margin-bottom: 10px;">
            <el-table-column
                    label="op"
                    width="60">
                <template slot="header">
                    <i type="primary" class="el-icon-plus" @click="newPortal"></i>
                </template>
                <template slot-scope="scope">
                    <i class="el-icon-delete" @click="delPortal(scope.$index)"></i>
                </template>
            </el-table-column>
            <el-table-column
                    prop="tag"
                    label="tag">
                <template slot-scope="scope">
                    <el-input v-model="scope.row.tag" placeholder="tag" v-setting></el-input>
                </template>
            </el-table-column>
            <el-table-column
                    prop="domain"
                    label="domain">
                <template slot-scope="scope">
                    <el-input v-model="scope.row.domain" placeholder="domain" v-setting></el-input>
                </template>
            </el-table-column>
        </el-table>
    </setting-card>
</template>

<script>
    export default {
        name: "Reverse",
        model: {
            prop: 'setting',
            event: 'change'
        },
        created() {
            this.fillDefaultValue(this.setting);
        },
        watch: {
            setting(val) {
                if(this.changedByForm){
                    this.changedByForm = false;
                    return;
                }

                this.fillDefaultValue(val);
            }
        },
        methods: {
            fillDefaultValue(setting) {
                setting = setting || {};
                Object.assign(this.sForm, setting);
                let bridges = this.sForm.bridges || [];
                let portals = this.sForm.portals || [];
                bridges.forEach(b => {
                    Object.assign({tag: "", domain: ""}, b );
                });
                portals.forEach(p => {
                    Object.assign({tag: "", domain: ""}, p);
                });
                this.sForm.portals = this._.cloneDeep(portals);
                this.sForm.bridges = this._.cloneDeep(bridges);
                this.$nextTick().then(()=>{
                    this.formChanged();
                });

            },
            newPortal() {
                this.sForm.portals.push({
                    "tag": "",
                    "domain": ""
                });
            },
            delPortal(idx) {
                this.sForm.portals.splice(idx, 1);
                this.formChanged();
            },
            newBridge() {
                this.sForm.bridges.push({
                    "tag": "",
                    "domain": ""
                });
            },
            delBridge(idx) {
                this.sForm.bridges.splice(idx, 1);
                this.formChanged();
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

                if (this.enableSetting) {
                    let portalTags = this.sForm.portals.map(p=>p.tag).filter(t=>t);
                    let bridgeTags = this.sForm.bridges.map(b=>b.tag).filter(t=>t);
                    this.$store.commit("setPortalTags", portalTags);
                    this.$store.commit("setBridgeTags", bridgeTags);
                } else {
                    this.$store.commit("setPortalTags", []);
                    this.$store.commit("setBridgeTags", []);
                }
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
                "enableSetting": true,
                changedByForm: false,
                sForm: {
                    bridges: [],
                    portals: [],
                }
            }
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
