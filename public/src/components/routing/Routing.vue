<template>
    <setting-card title="路由设置" :show-enable="false">
        <el-tooltip placement="top">
            <div slot="content">
                <ul>
                    <li>"AsIs"：只使用域名进行路由选择。默认值。</li>
                    <li>"IPIfNonMatch"：当域名没有匹配任何规则时，将域名解析成 IP（A 记录或 AAAA 记录）再次进行匹配；</li>
                    <ul>
                        <li>当一个域名有多个 A 记录时，会尝试匹配所有的 A 记录，直到其中一个与某个规则匹配为止；</li>
                        <li>解析后的 IP 仅在路由选择时起作用，转发的数据包中依然使用原始域名；</li>
                    </ul>
                    <li>"IPOnDemand"：当匹配时碰到任何基于 IP 的规则，将域名立即解析为 IP 进行匹配；</li>
                </ul>
            </div>
            <label>域名解析策略:</label>
        </el-tooltip>

        <el-select v-model="sForm.domainStrategy" v-setting style="width:200px;">
            <el-option value="AsIs"/>
            <el-option value="IPIfNonMatch"/>
            <el-option value="IPOnDemand"/>
        </el-select>
        <setting-card title="负载均衡器" :show-enable="false">
            <template v-slot:header-buttons>
                <i class="el-icon-plus" style="margin-right:10px;" @click="newBalancer"></i>
            </template>
            <el-form v-for="(b,idx) in sForm.balancers" label-width="80px" :key="idx">
                <el-row>
                    <el-col :span="8">
                        <el-form-item label="tag">
                            <el-input v-model="b.tag" v-setting>
                                <el-button slot="append" icon="el-icon-delete" @click="delBalancer(idx)"/>
                            </el-input>
                        </el-form-item>
                    </el-col>
                    <el-col :span="16">
                        <el-form-item label="selector">
                            <el-select :multiple="true" v-model="b.selector" filterable allow-create
                                       placeholder="outbound select" v-setting>
                                <el-option v-for="(tag,idx) in outboundTags" :value="tag" :key="idx"/>
                            </el-select>
                        </el-form-item>
                    </el-col>
                </el-row>
            </el-form>

        </setting-card>
        <setting-card title="规则列表" :show-enable="false">
            <template v-slot:header-buttons>
                <i class="el-icon-plus" style="margin-right:10px;" @click="newRule"></i>
            </template>
            <transition-group name="rule-list" tag="div">
                <Rule v-for="(rule,idx) in sForm.rules" :idx="idx" :rule="rule"
                      @new-rule="newRule" @del-rule="delRule" @copy-rule="copyRule" v-setting
                      :rules="sForm.rules" @change="sForm.rules[idx]=$event" :key="rule.id"/>
            </transition-group>
        </setting-card>

    </setting-card>
</template>

<script>
    import Rule from "@/components/routing/Rule";
    import {mapGetters} from "vuex";

    let ruleIdx = 0;
    let balancerIdx = 0;
    export default {
        name: "Routing",
        components: {Rule},
        model: {
            prop: "setting",
            event: "change"
        },
        data() {
            return {
                sForm: {
                    "domainStrategy": "AsIs",
                    "rules": [],
                    "balancers": []
                },
            }
        },
        props: {
            setting: {
                type: Object,
            }
        },
        computed: {
            ...mapGetters({"outboundTags": "getAllOutboundTags"}),
        },
        methods: {
            fillDefaultValue(setting) {
                setting = setting || {};
                //let oldRules = this.sForm.rules;
                //let oldBalancers = this.sForm.balancers;
                Object.assign(this.sForm, setting);
                let rules = this.sForm.rules || [];
                rules.forEach(rule => {
                    if (typeof (rule.id) !== "undefined" && rule.id > ruleIdx) {
                        ruleIdx = rule.id
                    }
                });
                let idMap = {}; // 用于防止id重复的情况
                rules.forEach(rule => {
                    if (typeof (rule.id) === "undefined" || typeof(idMap[rule.id]) !=="undefined") {
                        rule.id = ruleIdx++;
                    }
                    idMap[rule.id] = rule.id;
                });
                let balancers = this.sForm.balancers || [];
                balancers.forEach(balancer => {
                    if (typeof (balancer.id) !== "undefined" && balancer.id > balancerIdx) {
                        balancerIdx = balancer.id;
                    }
                });
                idMap = {}; // 用于防止id重复的情况
                balancers.forEach(balancer => {
                    if (typeof (balancer.id) === "undefined" || typeof(idMap[balancer.id]) !=="undefined") {
                        balancer.id = balancerIdx++;
                    }
                    idMap[balancer.id] = balancer.id;
                });
            },
            newRule(idx) {
                if (typeof (idx) !== "number") {
                    idx = this.sForm.rules.length - 1;
                }
                this.sForm.rules.splice(idx + 1, 0, {"id": ruleIdx++, "domain": [], ip: [], inboundTag: []});
            },
            copyRule(idx, setting) {
                if (typeof (idx) !== "number") {
                    idx = this.sForm.rules.length - 1;
                }
                setting.id = ruleIdx++;
                this.sForm.rules.splice(idx + 1, 0, setting);
            },
            delRule(idx) {
                this.sForm.rules.splice(idx, 1);
                this.formChanged();
            },
            newBalancer(idx) {
                if (typeof (idx) !== "number") {
                    idx = this.sForm.balancers.length - 1;
                }
                this.sForm.balancers.splice(idx + 1, 0, {"id": balancerIdx++});
            },
            delBalancer(idx) {
                this.sForm.balancers.splice(idx, 1);
                this.formChanged();
            },
            getSetting() {
                return Object.assign({}, this.sForm);
            },
            formChanged() {
                let balancersTag = this.sForm.balancers.map(b => b.tag).filter(t => t);
                this.$store.commit("setBalancersTag", balancersTag);
                this.changedByForm = true;
                this.$emit("change", this.getSetting());
            }
        },
        created() {
            this.fillDefaultValue(this.setting);
            this.formChanged();
        },
        watch: {
            setting(setting) {
                if (this.changedByForm) {
                    this.changedByForm = false;
                    return;
                }
                this.fillDefaultValue(setting);
            }
        }
    }
</script>

<style>
    .rule-list-move {
        transition: transform 1s;
    }

    /*.rule-list-item {*/
    /*    transition: all 1s;*/
    /*}*/
    /*.rule-list-enter, .rule-list-leave-to*/
    /*    !* .rule-list-leave-active for below version 2.1.8 *! {*/
    /*    opacity: 0;*/
    /*    transform: translateY(30px);*/
    /*}*/
    /*.rule-list-leave-active {*/
    /*    position: absolute;*/
    /*}*/
</style>
