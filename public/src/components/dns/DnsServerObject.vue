<template>
    <div class="vnext">
        <el-form :inline="false" label-width="120px" class="text-left">
            <el-row>
                <el-col :span="12">
                    <el-form-item label="DNS地址">
                        <el-input
                                v-model="sForm.address" v-setting>
                            <el-button slot="append" icon="el-icon-delete" @click="$emit('del-server', idx)"/>
                        </el-input>
                    </el-form-item>
                </el-col>
                <el-col :span="12">
                    <el-form-item label="DNS端口">
                        <el-input
                                v-model.number="sForm.port" v-setting>
                        </el-input>
                    </el-form-item>
                </el-col>
                <el-col :span="12">
                    <el-form-item label="domains">
                        <template v-slot:label>
                            <el-tooltip effect="light">
                                <div slot="content">
                                    <p>一个域名列表，此列表包含的域名，将优先使用此服务器进行查询。域名格式和 路由配置 中相同。</p>
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
                                <label>domains:</label>
                            </el-tooltip>
                        </template>
                        <el-input type="textarea" rows="3" v-model="domainValue" placeholder="domains"
                                  v-setting></el-input>
                    </el-form-item>
                </el-col>
                <el-col :span="12">
                    <el-form-item label="expectIPs:">
                        <template v-slot:label>
                            <el-tooltip effect="light">
                                <div slot="content">
                                    <p>（V2Ray 4.22.0+）一个 IP 范围列表，格式和 路由配置 中相同。</p>
                                    <p>当配置此项时，V2Ray DNS 会对返回的 IP 的进行校验，只返回包含 expectIPs 列表中的地址。</p>
                                    <p>如果未配置此项，会原样返回 IP 地址。</p>
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
                                <label>expectIPs:</label>
                            </el-tooltip>
                        </template>
                        <el-input type="textarea" rows="3" v-model="ipValue" placeholder="expectIPs" v-setting></el-input>
                    </el-form-item>
                </el-col>
            </el-row>



        </el-form>


    </div>
</template>

<script>


    export default {
        name: "DnsServerObject",
        model: {
            prop: 'setting',
            event: 'change'
        },
        methods: {

            getSettings() {
                if(this.sForm.port =="53" && this.sForm.domains.length==0 && this.sForm.expectIPs.length==0){
                    return "" + this.sForm.address;
                }
                return Object.assign({}, this.sForm);
            },
            formChanged() {
                this.changedByForm = true;
                this.$emit("change", this.getSettings());
            },
            fillDefaultValue(setting) {
                if(typeof(setting)==="string"){
                    setting = {"address": setting};
                }

                Object.assign(this.sForm, setting);
                this.$nextTick().then(() => {
                    this.formChanged();
                })
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
                    "address": "",
                    "port": 53,
                    "domains": [
                    ],
                    "expectIPs": [
                    ]
                }
            }
        },
        props: {
            setting: {
                default() {
                    return "";
                }
            },
            idx: {
                type: Number,
            }
        },
        computed: {
            domainValue: {
                get() {
                    if (!this.sForm.domains) {
                        return "";
                    }
                    return this.sForm.domains.join("\n");
                },
                set(val) {
                    if(val==""){
                        this.sForm.domains = [];
                        return;
                    }
                    this.sForm.domains = val.split("\n");
                }
            },
            ipValue: {
                get() {
                    if (!this.sForm.expectIPs) {
                        return "";
                    }
                    return this.sForm.expectIPs.join("\n");
                },
                set(val) {
                    if(val==""){
                        this.sForm.expectIPs = [];
                        return;
                    }
                    this.sForm.expectIPs = val.split("\n");
                }
            },
        },
    }
</script>

<style scoped>
    .vnext {
        margin-bottom: 20px;
        border-bottom: 1px solid #8c939d;
    }
</style>
