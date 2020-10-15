<template>
    <setting-card title="Domain Socket传输数据" :enable-setting.sync="enableSetting" @update:enableSetting="enableSettingChanged">
        <el-form :inline="true" label-width="90px" class="text-left">
            <el-form-item label="path">
                <template v-slot:label>
                    <el-tooltip placement="top" effect="light">
                        <div slot="content">一个合法的文件路径。在运行 V2Ray 之前，这个文件必须不存在。</div>
                        <label>path</label>
                    </el-tooltip>
                </template>
                <el-input v-model="sForm.path" v-setting/>
            </el-form-item>
            <el-form-item label="abstract" >
                <template v-slot:label>
                    <el-tooltip placement="top" effect="light">
                        <div slot="content">是否为 abstract domain socket，默认 false。</div>
                        <label>abstract</label>
                    </el-tooltip>
                </template>
                <el-switch v-model="sForm.abstract" v-setting/>
            </el-form-item>
        </el-form>
        <p>Domain Socket 使用标准的 Unix domain socket 来传输数据。它的优势是使用了操作系统内建的传输通道，而不会占用网络缓存。相比起本地环回网络（local loopback）来说，Domain socket 速度略快一些。</p>
        <p>目前仅可用于支持 Unix domain socket 的平台，如 macOS 和 Linux。在 Windows 上不可用。</p>
        <p>如果指定了 domain socket 作为传输方式，在入站出站代理中配置的端口和 IP 地址将会失效，所有的传输由 domain socket 取代。</p>
    </setting-card>
</template>

<script>
    export default {
        name: "DomainSocketObject",
        model: {
            prop: 'setting',
            event: 'change'
        },
        data() {
            return {
                enableSetting: false,
                changedByForm: false,
                sForm: {
                    "abstract": false,
                    "path": "/path/to/ds/file",
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

                this.formChanged();
            },

            getSettings() {
                if (!this.enableSetting) {
                   return null;
                }
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
