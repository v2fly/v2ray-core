<template>
    <setting-card title="出站协议配置" :showEnable="false">
        <el-row :gutter="10" class="outbound-header">
            <el-col :span="2" class="op-icons">
                <i type="success" :class="{'el-icon-arrow-down':!showAllDetail,'el-icon-arrow-up':showAllDetail}" circle @click="switchAllDetail"></i>
                <i type="primary" class="el-icon-plus" circle @click="newBound"></i>
            </el-col>
            <el-col :span="5">tag</el-col>
            <el-col :span="5">发送出口IP</el-col>
            <el-col :span="6">转发到Tag</el-col>
            <el-col :span="6">出站协议</el-col>
        </el-row>
        <Outbound ref="outbounds" v-for="(outbound,idx) of sForm.outbounds" :outbound="outbound"
                 @change="sForm.outbounds.splice(idx,1,$event)" v-setting :idx="idx" @to-top="toTop"
                 @new-bound="newBound"  @delete-bound="delBound" @copy-bound="copyBound" :key="outbound.id"/>
    </setting-card>
</template>

<script>
    import Outbound from "@/components/outbounds/Outbound";
    let outboundIdx = 1;
    export default {
        name: "Outbounds",
        components: {
            Outbound
        },
        model: {
            prop: 'outbounds',
            event: 'change'
        },
        data(){
            return {
                changedByForm:false,
                showAllDetail:false,
                sForm:{
                    outbounds:[]
                }
            }
        },
        created() {
            const outbounds = this.outbounds || [];
            for(let outbound in outbounds){
                this.sForm.outbounds.push(Object.assign({"id": outboundIdx++}, outbound));
            }
            this.$nextTick().then(()=>{
                this.formChanged();
            });
        },
        mounted() {
            // $(this.$el).on("change", "input", () => {
            //     this.formChanged();
            // });
        },
        watch: {
            outbounds: {
                handler: function (val) {
                    if(this.changedByForm){
                        this.changedByForm = false;
                        return;
                    }
                    let outbounds = [];
                    val.forEach((outbound)=>{
                        let oldInbound = this.getInboundByTag(outbound.tag);
                        let outboundId = oldInbound!=null ? oldInbound.id : outboundIdx++;
                        outbounds.push(Object.assign({id:outboundId}, outbound));
                    });

                    this.sForm.outbounds = outbounds;
                    this.$nextTick().then(()=>{
                        this.formChanged();
                    });
                },
                deep: false
            }
        },
        props: {
            outbounds: {
                type: Array,
                default() {
                    return []
                }
            }
        },
        methods: {
            formChanged() {
                let tags = this.sForm.outbounds.map(b=>b.tag).filter(t=>t);
                this.$store.commit("setOutboundTags", tags);
                this.changedByForm = true;
                this.$emit("change", this.getSettings());
            },
            getSettings() {
                let outbounds = JSON.parse(JSON.stringify(this.sForm.outbounds));
                outbounds.forEach(outbound=>{
                    delete outbound.id;
                });
                return outbounds;
            },
            getInboundByTag(tag) {
                return this._.find(this.sForm.outbounds, (outbound)=>outbound.tag === tag);
            },
            switchAllDetail(){
                if(this.$refs.outbounds){
                    this.$refs.outbounds.forEach(outbound=>{
                        outbound.$data.showDetail = !this.showAllDetail;
                    });
                }

                this.showAllDetail = !this.showAllDetail;
            },
            newBound(idx) {
                if(typeof(idx)!=="number"){
                    idx = this.sForm.outbounds.length - 1;
                }
                this.sForm.outbounds.splice(idx+1, 0, {
                    "id": outboundIdx++
                });
            },
            copyBound(idx, setting) {
                if (typeof (idx) !== "number") {
                    idx = this.sForm.outbounds.length - 1;
                }
                setting.id = outboundIdx++;
                setting.tag = setting.tag + "_copy"
                this.sForm.outbounds.splice(idx + 1, 0, setting);
            },
            delBound(idx) {
                this.sForm.outbounds.splice(idx, 1);
                this.formChanged();
            },
            toTop(idx) {
                let delItems = this.sForm.outbounds.splice(idx, 1);
                this.sForm.outbounds.splice(0, 0, ...delItems);
            }
        }
    }
</script>

<style>

.outbound-header {
    line-height: 40px;
}
</style>
