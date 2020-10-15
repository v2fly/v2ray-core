<template>
    <div>
        远端主机列表：<el-button type="primary" icon="el-icon-plus" size="small" @click="newServer">新增</el-button>
        <SocksServer v-for="(servers,idx) in sForm.servers" :setting="servers"
                    @del-server="delServer"
                    @change="sForm.servers.splice(idx, 1, $event)" v-setting
                    :idx="idx" :key="servers.id"/>
    </div>
</template>

<script>
    import SocksServer from "@/components/outbounds/SocksServer";
    let serversIdx = 0;
    export default {
        name: "OutboundSocksSetting",
        components:{SocksServer},
        model: {
            prop: 'setting',
            event: 'change'
        },
        methods:{
            newServer(){
                this.sForm.servers.push({
                    "id": serversIdx++
                });
            },
            delServer(idx) {
                this.sForm.servers.splice(idx, 1);
                this.formChanged();
            },
            getSettings() {
                return Object.assign({},this.sForm);
            },
            formChanged() {
                this.changedByForm = true;
                this.$emit("change", this.getSettings());
            },
            fillDefaultValue(setting) {
                setting = this.setting || {};
                let servers = setting.servers || [];


                let oldServers = this.sForm.servers || [];
                this.sForm.servers = [];
                servers.forEach((n) => {
                    let oldN = this._.find(oldServers, {"address":n.address, "port": n.port});
                    let idx = oldN? oldN.id: serversIdx++;
                    this.sForm.servers.push(Object.assign({"id": idx}, n));
                });
                if(this.sForm.servers.length==0){
                    this.sForm.servers.push({id:serversIdx++});
                }
                this.$nextTick().then(()=>{
                    this.formChanged();
                });
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
                    "servers": [{id: serversIdx++}]
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
