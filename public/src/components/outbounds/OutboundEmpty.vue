<template>
    <div>
        <p>该协议无特殊参数设置</p>
    </div>
</template>

<script>
    export default {
        name: "OutboundEmpty",
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
            fillDefaultValue() {
                // setting = this.setting || {};
                // setting = this._.pick(setting, ["network", "address", "port"]);
                // this.sForm = this._.defaults(setting, this.sForm);
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
