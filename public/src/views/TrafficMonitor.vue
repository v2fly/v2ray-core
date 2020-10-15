<template>
    <div>
        <el-form :inline="true" >
            <el-form-item label="自动刷新">
                <el-switch v-model="autoRefresh" @change="autoRefreshChanged"></el-switch>
            </el-form-item>
            <el-form-item>
                <el-input placeholder="输入关键字过滤" v-model="search"/>
            </el-form-item>

        </el-form>

        <el-table
                :data="counters.filter(data => !search || data.name.toLowerCase().includes(search.toLowerCase()))"
                :default-sort = "{prop: 'value', order: 'descending'}"
                style="width: 100%">
            <el-table-column
                    prop="name" sortable
                    label="name">
            </el-table-column>
            <el-table-column
                    prop="value" sortable
                    label="value" :formatter="trafficFormatter"
                    width="180">
            </el-table-column>
            <el-table-column
                    prop="rate" sortable
                    label="rate" :formatter="trafficFormatter"
                    width="180">
            </el-table-column>
        </el-table>
    </div>
</template>

<script>
    import Monitor from "@/api/Monitor";
    export default {
        name: "TrafficMonitor",
        data() {
            return {
                search:"",
                autoRefresh: true,
                bPauseRefresh: false,
                counters:[]
            }
        },
        created(){
            this.query();
            // this.startAutoRefresh();
        },
        destroyed(){
            if(this.queryInterval) {
                clearInterval(this.queryInterval);
                delete this.queryInterval;
            }
        },
        methods:{
            trafficFormatter(row, column, cellValue) {
                if(cellValue/1024<1){
                    return cellValue;
                }
                if(cellValue/1024/1024<1) {
                    return (cellValue/1024).toFixed(3) + "K";
                }
                if(cellValue/1024/1024/1024<1) {
                    return (cellValue/1024/1024).toFixed(3) + "M";
                }
                return (cellValue/1024/1024/1024).toFixed(3) + "G";
            },
            autoRefreshChanged() {
                if(this.autoRefresh && !this.bPauseRefresh) {
                    this.startAutoRefresh();
                }else{
                    this.stopAutoRefresh();
                }
            },
            pauseRefresh(bPauseRefresh) {
                this.bPauseRefresh = bPauseRefresh;
                this.autoRefreshChanged();
            },
            stopAutoRefresh() {
                if(this.queryInterval) {
                    clearInterval(this.queryInterval);
                    delete this.queryInterval;
                }
            },
            startAutoRefresh() {
                this.stopAutoRefresh();
                this.queryInterval = setInterval(()=>{
                    this.query();
                }, 1000);
            },
            async query() {
                let res = await Monitor.listCounters();
                if(res.bSuccess) {
                    this.counters = res.data.rows;
                }
            },

        },
    }
</script>

<style scoped>

</style>
