<template>
    <div class="rule-div rule-list-item">

        <el-form :inline="false" ref="form" label-width="110px">
            <div style="display: table;width:100%">
                <div style="display:table-cell;width:140px;">
                    <div style="line-height: 40px">
                        <i class="el-icon-plus" @click="$emit('new-rule', idx)"></i>
                        <i class="el-icon-delete" @click="$emit('del-rule', idx)"></i>
                        <i class="el-icon-copy-document" @click="$emit('copy-rule', idx, getSetting())"></i>
                        <ItemUpDownIcons :items="rules" :idx="idx"/>
                        <el-dropdown :hide-on-click="false" @command="ruleItemClick">
                            <i class="el-icon-menu"></i>
                            <el-dropdown-menu slot="dropdown">
                                <el-dropdown-item :icon="domainShow?'el-icon-circle-check':'el-icon-remove-outline'"
                                                  command="domain">domain
                                </el-dropdown-item>
                                <el-dropdown-item :icon="ipShow?'el-icon-circle-check':'el-icon-remove-outline'"
                                                  command="ip">
                                    ip
                                </el-dropdown-item>
                                <el-dropdown-item :icon="portShow?'el-icon-circle-check':'el-icon-remove-outline'"
                                                  command="port">port
                                </el-dropdown-item>
                                <el-dropdown-item :icon="networkShow?'el-icon-circle-check':'el-icon-remove-outline'"
                                                  command="network">network
                                </el-dropdown-item>
                                <el-dropdown-item :icon="sourceShow?'el-icon-circle-check':'el-icon-remove-outline'"
                                                  command="source">source
                                </el-dropdown-item>
                                <el-dropdown-item :icon="userShow?'el-icon-circle-check':'el-icon-remove-outline'"
                                                  command="user">user
                                </el-dropdown-item>
                                <el-dropdown-item :icon="inboundTagShow?'el-icon-circle-check':'el-icon-remove-outline'"
                                                  command="inboundTag">inboundTag
                                </el-dropdown-item>
                                <el-dropdown-item :icon="protocolShow?'el-icon-circle-check':'el-icon-remove-outline'"
                                                  command="protocol">protocol
                                </el-dropdown-item>
                                <el-dropdown-item :icon="attrsShow?'el-icon-circle-check':'el-icon-remove-outline'"
                                                  command="attrs">attrs
                                </el-dropdown-item>
                            </el-dropdown-menu>
                        </el-dropdown>
                    </div>
                </div>
                <div style="display: table-cell">
                    <el-row>
                        <el-col :span="8">
                            <el-form-item label="type:" label-width="50px">
                                <el-select v-model="sForm.type" v-setting>
                                    <el-option value="field"/>
                                </el-select>
                            </el-form-item>
                        </el-col>
                        <el-col :span="8">
                            <el-form-item label="outboundTag:">
                                <el-select v-model="sForm.outboundTag" placeholder="outboundTag" filterable allow-create
                                           v-setting>
                                    <el-option v-for="tag in outboundTags" :value="tag" :key="tag"/>
                                </el-select>
                            </el-form-item>
                        </el-col>
                        <el-col :span="8">
                            <el-form-item label="balancerTag:">
                                <el-select v-model="sForm.balancerTag" placeholder="balancerTag" v-setting>
                                    <el-option v-for="tag in balancersTag" :value="tag" :key="tag"/>
                                </el-select>
                            </el-form-item>
                        </el-col>

                    </el-row>
                </div>
            </div>

            <el-row class="rule-item-row">
                <el-col :span="12" v-if="domainShow"  @mouseover.native="ruleItemMouseEnter('domain')" @mouseout.native="ruleItemMouseLeave('domain')">
                    <i class="el-icon-remove-outline rule-item-close" @click="ruleItemClick('domain')" v-show="ruleItemCloseShow.domain"></i>
                    <el-form-item label="domain:">
                        <template v-slot:label>
                            <el-tooltip effect="light">
                                <div slot="content">
                                    <ul>
                                        <li>纯字符串：当此字符串匹配目标域名中任意部分，该规则生效。比如 "sina.com" 可以匹配 "sina.com"、"sina.com.cn" 和
                                            "www.sina.com"，但不匹配 "sina.cn"。
                                        </li>
                                        <li>正则表达式：由 <code>"regexp:"</code> 开始，余下部分是一个正则表达式。当此正则表达式匹配目标域名时，该规则生效。例如
                                            "regexp:\\.goo.*\\.com$" 匹配 "www.google.com"、"fonts.googleapis.com"，但不匹配
                                            "google.com"。
                                        </li>
                                        <li>子域名（推荐）：由 <code>"domain:"</code> 开始，余下部分是一个域名。当此域名是目标域名或其子域名时，该规则生效。例如
                                            "domain:v2ray.com" 匹配 "www.v2ray.com"、"v2ray.com"，但不匹配 "xv2ray.com"。
                                        </li>
                                        <li>完整匹配：由 <code>"full:"</code> 开始，余下部分是一个域名。当此域名完整匹配目标域名时，该规则生效。例如
                                            "full:v2ray.com"
                                            匹配 "v2ray.com" 但不匹配 "www.v2ray.com"。
                                        </li>
                                        <li>预定义域名列表：由 <code>"geosite:"</code> 开头，余下部分是一个名称，如 <code>geosite:google</code>
                                            或者
                                            <code>geosite:cn</code>。名称及域名列表参考 <a target="_blank"
                                                                                 href="https://www.v2fly.org/config/routing.html#%E9%A2%84%E5%AE%9A%E4%B9%89%E5%9F%9F%E5%90%8D%E5%88%97%E8%A1%A8">预定义域名列表</a>。
                                        </li>
                                        <li>从文件中加载域名：形如 <code>"ext:file:tag"</code>，必须以 <code>ext:</code>（小写）开头，后面跟文件名和标签，文件存放在
                                            <a target="_blank" href="https://www.v2fly.org/config/env.html#资源文件路径"
                                               class="">资源目录</a> 中，文件格式与
                                            <code>geosite.dat</code> 相同，标签必须在文件中存在。
                                        </li>
                                    </ul>
                                </div>
                                <label>domain:</label>
                            </el-tooltip>
                        </template>
                        <el-input type="textarea" rows="3" v-model="domainValue" placeholder="domain"

                                  v-setting></el-input>
                    </el-form-item>
                </el-col>
                <el-col :span="12" v-if="ipShow" @mouseover.native="ruleItemMouseEnter('ip')" @mouseout.native="ruleItemMouseLeave('ip')">
                    <i class="el-icon-remove-outline rule-item-close" @click="ruleItemClick('ip')" v-show="ruleItemCloseShow.ip"></i>
                    <el-form-item label="ip:">
                        <template v-slot:label>
                            <el-tooltip effect="light">
                                <div slot="content">
                                    <ul>
                                        <li>IP：形如 <code>"127.0.0.1"</code>。</li>
                                        <li><a href="https://en.wikipedia.org/wiki/Classless_Inter-Domain_Routing"
                                               target="_blank" rel="noopener noreferrer">CIDR
                                            <svg xmlns="http://www.w3.org/2000/svg" aria-hidden="true" x="0px" y="0px"
                                                 viewBox="0 0 100 100" width="15" height="15" class="icon outbound">
                                                <path fill="currentColor"
                                                      d="M18.8,85.1h56l0,0c2.2,0,4-1.8,4-4v-32h-8v28h-48v-48h28v-8h-32l0,0c-2.2,0-4,1.8-4,4v56C14.8,83.3,16.6,85.1,18.8,85.1z"></path>
                                                <polygon fill="currentColor"
                                                         points="45.7,48.7 51.3,54.3 77.2,28.5 77.2,37.2 85.2,37.2 85.2,14.9 62.8,14.9 62.8,22.9 71.5,22.9"></polygon>
                                            </svg>
                                        </a>：形如 <code>"10.0.0.0/8"</code>。
                                        </li>
                                        <li>GeoIP：形如 <code>"geoip:cn"</code>，必须以 <code>geoip:</code>（小写）开头，后面跟双字符国家代码，支持几乎所有可以上网的国家。
                                            <ul>
                                                <li>特殊值：<code>"geoip:private"</code>（V2Ray 3.5+），包含所有私有地址，如 <code>127.0.0.1</code>。
                                                </li>
                                            </ul>
                                        </li>
                                        <li>从文件中加载 IP：形如 <code>"ext:file:tag"</code>，必须以 <code>ext:</code>（小写）开头，后面跟文件名和标签，文件存放在
                                            <a href="/config/env.html#资源文件路径" class="">资源目录</a> 中，文件格式与
                                            <code>geoip.dat</code> 相同标签必须在文件中存在。
                                        </li>
                                    </ul>
                                </div>
                                <label>ip:</label>
                            </el-tooltip>
                        </template>
                        <el-input type="textarea" rows="3" v-model="ipValue" placeholder="ip" v-setting></el-input>
                    </el-form-item>
                </el-col>
                <el-col :span="12" v-if="portShow" @mouseover.native="ruleItemMouseEnter('port')" @mouseout.native="ruleItemMouseLeave('port')">
                    <i class="el-icon-remove-outline rule-item-close" @click="ruleItemClick('port')" v-show="ruleItemCloseShow.port"></i>
                    <el-form-item label="port:">
                        <el-input v-model="sForm.port" placeholder="port" v-setting></el-input>
                    </el-form-item>
                </el-col>
                <el-col :span="12" v-if="networkShow" @mouseover.native="ruleItemMouseEnter('network')" @mouseout.native="ruleItemMouseLeave('network')">
                    <i class="el-icon-remove-outline rule-item-close" @click="ruleItemClick('network')" v-show="ruleItemCloseShow.network"></i>
                    <el-form-item label="network:">
                        <el-select v-model="sForm.network" placeholder="network" v-setting>
                            <el-option value="tcp"/>
                            <el-option value="udp"/>
                            <el-option value="tcp,udp"/>
                        </el-select>
                    </el-form-item>
                </el-col>
                <el-col :span="12" v-if="sourceShow" @mouseover.native="ruleItemMouseEnter('source')" @mouseout.native="ruleItemMouseLeave('source')">
                    <i class="el-icon-remove-outline rule-item-close" @click="ruleItemClick('source')" v-show="ruleItemCloseShow.source"></i>
                    <el-form-item label="source:">
                        <el-input v-model="sourceValue" placeholder="source" v-setting></el-input>
                    </el-form-item>
                </el-col>
                <el-col :span="12" v-if="userShow" @mouseover.native="ruleItemMouseEnter('user')" @mouseout.native="ruleItemMouseLeave('user')">
                    <i class="el-icon-remove-outline rule-item-close" @click="ruleItemClick('user')" v-show="ruleItemCloseShow.user"></i>
                    <el-form-item label="user:">
                        <el-input v-model="userValue" placeholder="user" v-setting></el-input>
                    </el-form-item>
                </el-col>
                <el-col :span="12" v-if="inboundTagShow" @mouseover.native="ruleItemMouseEnter('inboundTag')" @mouseout.native="ruleItemMouseLeave('inboundTag')">
                    <i class="el-icon-remove-outline rule-item-close" @click="ruleItemClick('inboundTag')" v-show="ruleItemCloseShow.inboundTag"></i>
                    <el-form-item label="inboundTag:">
                        <el-select :multiple="true" v-model="sForm.inboundTag" filterable allow-create
                                   placeholder="inboundTag" v-setting>
                            <el-option v-for="(tag,idx) in inboundTags" :value="tag" :key="idx"/>
                        </el-select>
                    </el-form-item>
                </el-col>
                <el-col :span="12" v-if="protocolShow" @mouseover.native="ruleItemMouseEnter('protocol')" @mouseout.native="ruleItemMouseLeave('protocol')">
                    <i class="el-icon-remove-outline rule-item-close" @click="ruleItemClick('protocol')" v-show="ruleItemCloseShow.protocol"></i>
                    <el-form-item label="protocol:">
                        <el-select :multiple="true" v-model="sForm.protocol" placeholder="protocol" v-setting>
                            <el-option value="http"/>
                            <el-option value="tls"/>
                            <el-option value="bittorrent"/>
                        </el-select>
                    </el-form-item>
                </el-col>
                <el-col :span="12" v-if="attrsShow" @mouseover.native="ruleItemMouseEnter('attrs')" @mouseout.native="ruleItemMouseLeave('attrs')">
                    <i class="el-icon-remove-outline rule-item-close" @click="ruleItemClick('attrs')" v-show="ruleItemCloseShow.attrs"></i>
                    <el-form-item label="attrs:">
                        <el-input v-model="sForm.attrs" placeholder="attrs" v-setting/>
                    </el-form-item>
                </el-col>
            </el-row>
        </el-form>


    </div>
</template>

<script>
    import ItemUpDownIcons from "@/components/ItemUpDownIcons";
    import {mapGetters} from 'vuex'

    export default {
        name: "Rule",
        components: {ItemUpDownIcons},
        model: {
            prop: "rule",
            event: "change"
        },
        props: {
            rule: {
                type: Object,
            },
            rules: {
                type: Array,
            },
            idx: {
                type: Number,
            }
        },
        methods: {
            ruleItemMouseEnter(itemName) {
                this.ruleItemCloseShow[itemName] = true;
            },
            ruleItemMouseLeave(itemName) {
                this.ruleItemCloseShow[itemName] = false;
            },
            ruleItemClick(itemName) {
                let isArrayValue = true;
                let isShow = true;
                if (this._.indexOf(["port", "network", "attrs"], itemName) != -1) {
                    isArrayValue = false;
                }
                if (typeof (this.sForm[itemName]) === "undefined" || this.sForm[itemName] === null) {
                    isShow = false;
                }
                if (isShow) {
                    this.sForm[itemName] = null;
                    this.ruleItemCloseShow[itemName] = false;
                } else {

                    if (isArrayValue) {
                        this.$set(this.sForm, itemName, []);
                    } else {
                        this.$set(this.sForm, itemName, "");
                    }
                }
                this.formChanged();
            },
            getSetting() {
                return this._.cloneDeep(this.sForm);
            },
            formChanged() {
                this.changedByForm = true;
                this.$emit("change", this.getSetting());
            }
        },
        computed: {
            ...mapGetters({
                "inboundTags": "getAllInboundTags",
                "outboundTags": "getAllOutboundTags",
                "balancersTag": "getBalancersTag"
            }),
            domainShow() {
                if (typeof (this.sForm.domain) === 'undefined') {
                    return false;
                }
                if (this.sForm.domain === null) {
                    return false;
                }
                return true;
            },
            domainValue: {
                get() {
                    if (!this.sForm.domain) {
                        return "";
                    }
                    return this.sForm.domain.join("\n");
                },
                set(val) {
                    this.sForm.domain = val.split("\n");
                }
            },
            ipShow() {
                if (typeof (this.sForm.ip) === 'undefined') {
                    return false;
                }
                if (this.sForm.ip === null) {
                    return false;
                }
                return true;
            },
            ipValue: {
                get() {
                    if (!this.sForm.ip) {
                        return "";
                    }
                    return this.sForm.ip.join("\n");
                },
                set(val) {
                    this.sForm.ip = val.split("\n");
                }
            },
            sourceShow() {
                if (typeof (this.sForm.source) === 'undefined') {
                    return false;
                }
                if (this.sForm.source === null) {
                    return false;
                }
                return true;
            },
            sourceValue: {
                get() {
                    if (!this.sForm.source) {
                        return "";
                    }
                    return this.sForm.source.join(",");
                },
                set(val) {
                    this.sForm.source = val.split(",");
                }
            },
            userShow() {
                if (typeof (this.sForm.user) === 'undefined') {
                    return false;
                }
                if (this.sForm.user === null) {
                    return false;
                }
                return true;
            },
            userValue: {
                get() {
                    if (!this.sForm.user) {
                        return "";
                    }
                    return this.sForm.user.join(",");
                },
                set(val) {
                    this.sForm.user = val.split(",");
                }
            },
            inboundTagShow() {
                if (typeof (this.sForm.inboundTag) === 'undefined') {
                    return false;
                }
                if (this.sForm.inboundTag === null) {
                    return false;
                }
                return true;
            },
            inboundTagValue: {
                get() {
                    if (!this.sForm.inboundTag) {
                        return "";
                    }
                    return this.sForm.inboundTag.join(",");
                },
                set(val) {
                    this.sForm.inboundTag = val.split(",");
                }
            },
            portShow() {
                if (typeof (this.sForm.port) === 'undefined') {
                    return false;
                }
                if (this.sForm.port === null) {
                    return false;
                }
                return true;
            },
            attrsShow() {
                if (typeof (this.sForm.attrs) === 'undefined') {
                    return false;
                }
                if (this.sForm.attrs === null) {
                    return false;
                }
                return true;
            },
            networkShow() {
                if (typeof (this.sForm.network) === 'undefined') {
                    return false;
                }
                if (this.sForm.network === null) {
                    return false;
                }
                return true;
            },
            protocolShow() {
                if (typeof (this.sForm.protocol) === 'undefined') {
                    return false;
                }
                if (this.sForm.protocol === null) {
                    return false;
                }
                return true;
            }
        },
        data() {
            return {
                changedByForm: false,
                ruleItemCloseShow: {
                    domain: false,
                    ip: false,
                    port: false,
                    sourcePort: false,
                    network: false,
                    source: false,
                    user: false,
                    inboundTag: false,
                    protocol: false,
                    attrs: false,
                },
                sForm: {
                    "type": "field",
                    "outboundTag": "direct",
                    "balancerTag": "",
                }
            }
        },
        created() {
            this.sForm = Object.assign({}, this.sForm, this._.cloneDeep(this.rule));
            this.formChanged();
        },
        watch: {
            rule(rule) {
                this.sForm = Object.assign({}, this.sForm, this._.cloneDeep(rule));
            }
        }
    }
</script>

<style>
    .rule-div {
        border-bottom: 1px solid #8c939d;
        margin-bottom: 10px;
    }

    .rule-item-row > div {
        position: relative;
    }

    .rule-item-close {
        position: absolute;
        right: 0;
        top: 0;
        z-index: 99;
        color: red;
        cursor: pointer;
    }
</style>
