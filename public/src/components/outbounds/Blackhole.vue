<template>
    <div>
        <el-form :inline="false" label-width="110px" class="text-left">
            <el-row :gutter="10">
                <el-col :span="8">
                    <el-form-item label="responseType">
                        <el-select v-model="sForm.response.type" v-setting>
                            <el-option value="none"/>
                            <el-option value="http"/>
                        </el-select>
                    </el-form-item>
                </el-col>
            </el-row>
        </el-form>
        <p>Blackhole（黑洞）是一个出站数据协议，它会阻碍所有数据的出站，配合 <a href="https://www.v2fly.org/config/routing.html" target="_blank">路由（Routing）</a>
            一起使用，可以达到禁止访问某些网站的效果。</p>
        <p>当 responseType 为 "none"（默认值）时，Blackhole 将直接关闭连接。当 type 为 "http" 时，Blackhole 会发回一个简单的 HTTP 403 数据包，然后关闭连接。</p>
    </div>
</template>

<script>
    export default {
        name: "Blackhole",
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
                setting = this._.pick(setting, ["response"]);
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
                    "response": {
                        "type": "none"
                    }
                }
            }
        },
        props: {
            setting: {
                type: Object
            }
        },
        computed: {},
    }
</script>

<style scoped>

</style>
