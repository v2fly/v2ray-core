<template>
    <setting-card title="DNS服务器配置" :show-enable="true"
                  @update:enableSetting="enableSettingChanged"
                  :enableSetting.sync="enableSetting">
        <el-form ref="form" :model="sForm" label-width="80px">
            <el-row>
                <el-col :span="12">
                    <el-form-item label="tag">
                        <el-input v-model="sForm.tag" v-setting></el-input>
                    </el-form-item>
                </el-col>
                <el-col :span="12">
                    <el-form-item label="clientIp">
                        <el-tooltip slot="label" effect="light">
                            <div slot="content">当前系统的 IP 地址，用于 DNS 查询时，通知服务器客户端的所在位置。不能是私有地址。</div>
                            <label>clientIp</label>
                        </el-tooltip>
                        <el-input v-model="sForm.clientIp" v-setting></el-input>
                    </el-form-item>
                </el-col>
                <el-col :span="24">
                    <el-form-item label="hosts">
                        <el-tooltip effect="light" slot="label">
                            <div slot="content">
                                <p>静态 IP 列表，每行是格式为"域名 地址", 例如："www.baidu.com 127.0.0.1"</p>
                                <p>其中地址可以是 IP 或者域名。在解析域名时，如果域名匹配这个列表中的某一项，当该项的地址为 IP 时，则解析结果为该项的 IP，而不会使用下述的 servers
                                    进行解析；</p>
                                <p>当该项的地址为域名时，会使用此域名进行 IP 解析，而不使用原始域名。</p>
                                <p>域名的格式有以下几种形式：</p>
                                <ul>
                                    <li>纯字符串：当此域名完整匹配目标域名时，该规则生效。例如 "v2ray.com" 匹配 "v2ray.com" 但不匹配 "www.v2ray.com"。
                                    </li>
                                    <li>正则表达式：由 <code>"regexp:"</code> 开始，余下部分是一个正则表达式。当此正则表达式匹配目标域名时，该规则生效。例如
                                        "regexp:\\.goo.*\\.com$" 匹配 "www.google.com"、"fonts.googleapis.com"，但不匹配
                                        "google.com"。
                                    </li>
                                    <li>子域名 (推荐)：由 <code>"domain:"</code> 开始，余下部分是一个域名。当此域名是目标域名或其子域名时，该规则生效。例如
                                        "domain:v2ray.com" 匹配 "www.v2ray.com"、"v2ray.com"，但不匹配 "xv2ray.com"。
                                    </li>
                                    <li>子串：由 <code>"keyword:"</code> 开始，余下部分是一个字符串。当此字符串匹配目标域名中任意部分，该规则生效。比如
                                        "keyword:sina.com" 可以匹配 "sina.com"、"sina.com.cn" 和 "www.sina.com"，但不匹配
                                        "sina.cn"。
                                    </li>
                                    <li>预定义域名列表：由 <code>"geosite:"</code> 开头，余下部分是一个名称，如 <code>geosite:google</code> 或者
                                        <code>geosite:cn</code>。名称及域名列表参考 <a href="/config/routing.html#dlc" class="">预定义域名列表</a>。
                                    </li>
                                </ul>
                            </div>
                            <label>主机列表</label>
                        </el-tooltip>
                        <div class="el-textarea">
                                <textarea autocomplete="off" rows="5" :value="hostsValue"
                                          @change="hostValueChange($event)" v-setting
                                          placeholder="domain address"
                                          class="el-textarea__inner"
                                          style="min-height: 33px;"></textarea>
                        </div>
                    </el-form-item>
                </el-col>
            </el-row>
        </el-form>

        <p>
            DNS解析服务器：
            <el-button type="primary" icon="el-icon-plus" size="small" @click="newServer">新增</el-button>
        </p>
        <DnsServerObject v-for="(servers,idx) in sForm.servers" :setting="servers"
                           @del-server="delServer"
                           @change="sForm.servers.splice(idx, 1, $event)" v-setting
                           :idx="idx" :key="idx"/>

    </setting-card>
</template>

<script>
    import DnsServerObject from "@/components/dns/DnsServerObject";

    export default {
        name: "DnsSetting",
        model: {
            prop: 'setting',
            event: 'change'
        },
        components: {DnsServerObject},
        created() {
            const setting = this.setting || {};
            Object.assign(this.sForm, setting);
            this.formChanged();
        },
        mounted() {


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
        methods: {
            newServer(){
                this.sForm.servers.push({
                });
            },
            delServer(idx) {
                this.sForm.servers.splice(idx, 1);
            },
            hostValueChange(event) {
                let val = event.currentTarget.value;
                val = val.trim();
                let hosts = val.split(/[\r\n]+/);
                this.hosts = [];
                hosts.forEach(h => {
                    let domainIp = h.trim().split(/ +/);
                    if (domainIp.length < 2) {
                        return;
                    }
                    this.hosts.push({"domain": domainIp[0], "address": domainIp[1]});
                });
            },
            fillDefaultValue(setting) {
                setting = setting || {};
                if(setting.tag){
                    this.enableSetting = true;
                }else{
                    this.enableSetting = false;
                }
                Object.assign(this.sForm, setting);
                let hosts = this.sForm.hosts || {};
                this.hosts = [];
                for (let domain in hosts) {
                    this.hosts.push({domain: domain, address: hosts[domain]});
                }

                this.$nextTick().then(() => {
                    this.formChanged();
                })

            },
            getSettings() {
                if (!this.enableSetting) {
                    return null;
                }
                this.sForm.hosts = {};
                this.hosts.forEach(h => {
                    this.sForm.hosts[h.domain] = h.address;
                });
                return Object.assign({}, this.sForm);
            },
            enableSettingChanged() {
                this.$nextTick(() => {
                    this.formChanged();
                });

            },
            formChanged() {

                if (this.enableSetting) {
                    this.$store.commit("setDnsTag", this.sForm.tag);
                } else {
                    this.$store.commit("setDnsTag", null);
                }
                let newSetting = this.getSettings();
                if(newSetting !== this.setting) {
                    this.changedByForm = true;
                    this.$emit("change", newSetting);
                }

            }
        },
        computed: {
            hostsValue: {
                set(val) {
                    val = val.trim();
                    let hosts = val.split(/[\r\n]+/);
                    this.hosts = [];
                    hosts.forEach(h => {
                        let domainIp = h.trim().split(/ +/);
                        if (domainIp.length < 2) {
                            return;
                        }
                        this.hosts.push({"domain": domainIp[0], "address": domainIp[1]});
                    })
                },
                get() {
                    let hosts = [];
                    this.hosts.forEach(h => {
                        hosts.push(h.domain + " " + h.address);
                    });
                    return hosts.join("\n");
                }
            }
        },
        data() {
            return {
                sForm: {
                    "hosts": {

                    },
                    "servers": [],
                    "clientIp": "1.2.3.4",
                    "tag": "dns_inbound"
                },
                hosts: [],
                servers: [],
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
