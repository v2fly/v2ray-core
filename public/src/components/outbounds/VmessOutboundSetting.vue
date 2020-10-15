<template>
    <div>
        远端主机列表：<el-button type="primary" icon="el-icon-plus" size="small" @click="newVnext">新增</el-button>
        <VmessVnext v-for="(vnext,idx) in sForm.vnext" :setting="vnext"
                    @del-vnext="delVnext"
                    @change="sForm.vnext.splice(idx, 1, $event)" v-setting
                    :idx="idx" :key="vnext.id"/>
    </div>
</template>

<script>
    import VmessVnext from "@/components/outbounds/VmessVnext";
    let vnextIdx = 0;
    export default {
        name: "VmessOutboundSetting",
        components:{VmessVnext},
        model: {
            prop: 'setting',
            event: 'change'
        },
        methods:{
            newVnext(){
                this.sForm.vnext.push({
                    "id": vnextIdx++
                });
            },
            delVnext(idx) {
                this.sForm.vnext.splice(idx, 1);
                this.formChanged();
            },
            getSettings() {
                let setting = this._.cloneDeep(this.sForm);
                let vnext = setting.vnext || [];
                vnext.forEach(v=>{
                    delete v.id;
                });
                return setting;
            },
            formChanged() {
                this.changedByForm = true;
                this.$emit("change", this.getSettings());
            },
            fillDefaultValue(setting) {
                setting = this.setting || {};
                let vnext = setting.vnext || [];


                let oldVnext = this.sForm.vnext || [];
                this.sForm.vnext = [];
                vnext.forEach((n) => {
                    let oldN = this._.find(oldVnext, {"address":n.address, "port": n.port});
                    let idx = oldN? oldN.id: vnextIdx++;
                    this.sForm.vnext.push(Object.assign({"id": idx}, n));
                });
                if(this.sForm.vnext.length==0){
                    this.sForm.vnext.push({id:vnextIdx++});
                }
                this.$nextTick().then(()=>{
                    this.formChanged();
                })
            }
        },
        watch:{
            setting(val){
                if(this.changedByForm){
                    this.changedByForm = false;
                    return;
                }
                this.fillDefaultValue(val);
            }
        },

        created() {
            this.fillDefaultValue(this.setting);
        },
        mounted(){
            // $(this.$el).on("change", "input", ()=>{
            //     this.formChanged();
            // });
        },

        data(){
            return {
                changedByForm: false,
                sForm:{
                    "vnext": [{id: vnextIdx++}]
                }
            }
        },
        props:{
            setting: {
                type: Object
            }
        },
        computed:{
            clientDefault:{
                get(){
                    return this.sForm.default;
                },
                set(newDef){
                    this.sForm.default = newDef;
                }
            }
        },
    }
</script>

<style scoped>

</style>
