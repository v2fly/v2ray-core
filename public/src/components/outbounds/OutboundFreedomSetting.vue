<template>
    <div>


        <el-form :inline="false" label-width="110px" class="text-left">
            <el-row :gutter="10">
                <el-col :span="8">
                    <el-form-item label="domainStrategy">
                        <el-select v-model="sForm.domainStrategy" v-setting>
                            <el-option label="AsIs" value="AsIs"></el-option>
                            <el-option label="UseIP" value="UseIP"></el-option>
                            <el-option label="UseIPv4" value="UseIPv4"></el-option>
                            <el-option label="UseIPv6" value="UseIPv6"></el-option>
                        </el-select>
                    </el-form-item>
                </el-col>
                <el-col :span="8">
                    <el-form-item label="redirect">
                        <el-input v-model="sForm.redirect" v-setting/>
                    </el-form-item>
                </el-col>
                <el-col :span="8">
                    <el-form-item label="userLevel">
                        <el-input-number v-model="sForm.userLevel" v-setting/>
                    </el-form-item>
                </el-col>
            </el-row>
        </el-form>
    </div>
</template>

<script>
    export default {
        name: "OutboundFreedomSetting",
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
                setting = this._.pick(setting, ["domainStrategy", "redirect", "userLevel"]);
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
                    "domainStrategy": "AsIs",
                    "redirect": "",
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
