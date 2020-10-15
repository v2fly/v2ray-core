<template>
    <div>
        <el-form :inline="true">
            <el-form-item label="自动加载">
                <el-switch v-model="autoRefresh" @change="autoRefreshChanged"></el-switch>
            </el-form-item>
            <el-form-item>
                <el-button @click="contents=[];from=0;loadLogContent()">头部</el-button>
                <el-button @click="contents=[];from=-1;loadLogContent()">尾部</el-button>
                <el-button @click="loadLogContent">加载更多</el-button>
            </el-form-item>
        </el-form>
        <pre v-for="(c,idx) in contents" :key="idx">{{c}}</pre>
    </div>
</template>

<script>
    import Log from "@/api/Log";

    export default {
        name: "LogViewer",
        data() {
            return {
                from: -1,
                contents: [],
                autoRefresh: true,
                bPauseRefresh: false,
            }
        },
        created() {
            this.loadLogContent();
            // this.startAutoRefresh();
        },
        destroyed() {
            if (this.queryInterval) {
                clearInterval(this.queryInterval);
                delete this.queryInterval;
            }
        },
        methods: {
            autoRefreshChanged() {
                if (this.autoRefresh && !this.bPauseRefresh) {
                    this.startAutoRefresh();
                } else {
                    this.stopAutoRefresh();
                }
            },
            pauseRefresh(bPauseRefresh) {
                this.bPauseRefresh = bPauseRefresh;
                this.autoRefreshChanged();
            },
            stopAutoRefresh() {
                if (this.queryInterval) {
                    clearInterval(this.queryInterval);
                    delete this.queryInterval;
                }
            },
            startAutoRefresh() {
                this.stopAutoRefresh();
                this.queryInterval = setInterval(() => {
                    this.loadLogContent();
                }, 1000);
            },
            async loadLogContent() {
                let res = await Log.loadLogContent({logType: this.logType, from: this.from})
                if (res.bSuccess) {
                    if(res.data.content){
                        this.contents.push(res.data.content);
                    }
                    this.from = res.data.lastPos;
                } else {
                    this.$message.error(res.msg);
                }
            }
        },
        props: {
            logType: {
                type: String,
                default: "access",
            },
        }
    }
</script>

<style scoped>
    pre {
        margin: 0;
    }
</style>
