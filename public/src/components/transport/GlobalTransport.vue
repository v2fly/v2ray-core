<template>
    <setting-card title="全局传输设置" :show-enable="true"
                  @update:enableSetting="enableSettingChanged"
                  :enableSetting.sync="enableSetting">
        <TcpObject v-setting v-model="sForm.tcpSettings" />
        <WebSocketObject v-setting v-model="sForm.wsSettings" />
        <Http2Object v-setting v-model="sForm.httpSettings" />
        <DomainSocketObject v-setting v-model="sForm.dsSettings" />
        <QuicObject v-setting v-model="sForm.quicSettings" />
        <KcpObject v-setting v-model="sForm.kcpSettings" />
    </setting-card>
</template>

<script>
    import WebSocketObject from "@/components/transport/WebSocketObject";
    import Http2Object from "@/components/transport/Http2Object";
    import DomainSocketObject from "@/components/transport/DomainSocketObject";
    import TcpObject from "@/components/transport/TcpObject";
    import QuicObject from "@/components/transport/QuicObject";
    import KcpObject from "@/components/transport/KcpObject";
    export default {
        name: "GlobalTransport",
        components: {WebSocketObject, Http2Object, DomainSocketObject, TcpObject, QuicObject, KcpObject},
        model: {
            prop: 'setting',
            event: 'change'
        },
        created() {
            const setting = this.setting || {};
            Object.assign(this.sForm, setting);
            this.formChanged();
        },
        mounted() {
        },
        watch: {
            setting(val) {
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

                this.changedByForm = true;
                this.$emit("change", this.getSettings());
            }
        },
        data() {
            return {
                sForm: {
                    "tcpSettings": {},
                    "kcpSettings": {},
                    "wsSettings": {},
                    "httpSettings": {},
                    "dsSettings": {},
                    "quicSettings": {}
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
