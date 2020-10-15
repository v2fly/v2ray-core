<template>
    <setting-card title="入站协议配置" :showEnable="false">
        <el-row :gutter="10" class="inbound-header">
            <el-col :span="2" class="op-icons">
                <i type="success" :class="{'el-icon-arrow-down':!showAllDetail,'el-icon-arrow-up':showAllDetail}" circle
                   @click="switchAllDetail"></i>
                <i type="primary" class="el-icon-plus" circle @click="newBound"></i>
            </el-col>
            <el-col :span="5">tag</el-col>
            <el-col :span="5">监听地址</el-col>
            <el-col :span="6">
                <el-tooltip effect="light" placement="top">
                    <div slot="content">
                        <p>端口。接受的格式如下:</p>
                        <ul>
                            <li>整型数值：实际的端口号。</li>
                            <li>环境变量：以 <code>"env:"</code> 开头，后面是一个环境变量的名称，如 <code>"env:PORT"</code>。V2Ray
                                会以字符串形式解析这个环境变量。
                            </li>
                            <li>字符串：可以是一个数值类型的字符串，如 <code>"1234"</code>；或者一个数值范围，如 <code>"5-10"</code> 表示端口 5 到端口 10，这 6
                                个端口。
                            </li>
                        </ul>
                        <p>当只有一个端口时，V2Ray 会在此端口监听入站连接。当指定了一个端口范围时，取决于 <code>allocate</code> 设置。</p>
                    </div>
                    <label>监听端口</label>
                </el-tooltip>
            </el-col>
            <el-col :span="6">监听协议</el-col>
        </el-row>
        <Inbound ref="inbounds" v-for="(inbound,idx) of sForm.inbounds" :inbound="inbound"
                 @change="sForm.inbounds.splice(idx, 1, $event)" v-setting :idx="idx"
                 @new-inbound="newBound" @delete-inbound="delBound" @copy-inbound="copyBound" :key="inbound.id"/>
    </setting-card>
</template>

<script>
    import Inbound from "@/components/inbounds/Inbound";

    let inboundIdx = 1;
    export default {
        name: "Inbounds",
        components: {
            Inbound
        },
        model: {
            prop: 'inbounds',
            event: 'change'
        },
        data() {
            return {
                changedByForm: false,
                showAllDetail: false,
                sForm: {
                    inbounds: []
                }
            }
        },
        created() {
            const inbounds = this.inbounds || [];
            for (let inbound in inbounds) {
                this.sForm.inbounds.push(Object.assign({"id": inboundIdx++}, inbound));
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
            inbounds: {
                handler: function (val) {
                    if (this.changedByForm) {
                        this.changedByForm = false;
                        return;
                    }
                    let inbounds = [];
                    val.forEach((inbound) => {
                        let oldInbound = this.getInboundByTag(inbound.tag);
                        let inboundId = oldInbound != null ? oldInbound.id : inboundIdx++;
                        inbounds.push(Object.assign({id: inboundId}, inbound));
                    });

                    this.sForm.inbounds = inbounds;

                    this.$store.commit('setInboundTags', this.sForm.inbounds.filter(inbound => inbound.tag).map(inbound => {
                        return inbound.tag;
                    }));

                    this.$nextTick().then(()=>{
                        this.formChanged();
                    });
                },
                deep: false
            }
        },
        props: {
            inbounds: {
                type: Array,
                default() {
                    return []
                }
            }
        },
        methods: {
            formChanged() {
                this.changedByForm = true;
                this.$emit("change", this.getSettings());
                this.$store.commit('setInboundTags', this.sForm.inbounds.filter(inbound => inbound.tag).map(inbound => {
                    return inbound.tag;
                }));
            },
            getSettings() {
                let inbounds = JSON.parse(JSON.stringify(this.sForm.inbounds));
                inbounds.forEach(inbound => {
                    delete inbound.id;
                });
                return inbounds;
            },
            getInboundByTag(tag) {
                return this._.find(this.sForm.inbounds, (inbound) => inbound.tag === tag);
            },
            switchAllDetail() {
                if (this.$refs.inbounds) {
                    this.$refs.inbounds.forEach(inbound => {
                        inbound.$data.showDetail = !this.showAllDetail;
                    });
                }

                this.showAllDetail = !this.showAllDetail;
            },
            newBound(idx) {
                if (typeof (idx) !== "number") {
                    idx = this.sForm.inbounds.length - 1;
                }
                this.sForm.inbounds.splice(idx + 1, 0, {
                    "id": inboundIdx++
                });
            },
            copyBound(idx, setting) {
                if (typeof (idx) !== "number") {
                    idx = this.sForm.inbounds.length - 1;
                }
                setting.id = inboundIdx++;
                setting.tag = setting.tag + "_copy"
                this.sForm.inbounds.splice(idx + 1, 0, setting);
            },
            delBound(idx) {
                this.sForm.inbounds.splice(idx, 1);
                this.formChanged();
            }
        }
    }
</script>

<style>


    .inbound-header {
        line-height: 40px;
    }
</style>
