<template>
    <div>
        <el-form :inline="false" label-width="110px" class="text-left">
            <el-row :gutter="10">
                <el-col :span="8">
                    <el-form-item label="network">
                        <el-select v-model="sForm.network" v-setting>
                            <el-option value="tcp"/>
                            <el-option value="udp"/>
                        </el-select>
                    </el-form-item>
                </el-col>
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
            </el-row>
        </el-form>
        <p>DNS 是一个出站协议，主要用于拦截和转发 DNS 查询。此出站协议只能接收 DNS 流量（包含基于 UDP 和 TCP 协议的查询），其它类型的流量会导致错误。</p>
        <p>在处理 DNS 查询时，此出站协议会将 IP 查询（即 A 和 AAAA）转发给内置的 <a href="https://www.v2fly.org/config/dns.html" class="">DNS 服务器</a>。其它类型的查询流量将被转发至它们原本的目标地址。</p>
        <p>DNS 出站协议在 V2Ray 4.15 中引入。</p>
    </div>
</template>

<script>
    export default {
        name: "OutboundDNS",
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
                setting = this._.pick(setting, ["network", "address", "port"]);
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
                    "network": "tcp",
                    "address": "8.8.8.8",
                    "port": 53
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
