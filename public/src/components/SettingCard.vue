<template>

    <el-card class="setting-card" :body-style="bodyStyle">
        <div slot="header" class="clearfix" @dblclick="showSettingClick">
            <span>{{title}}</span>
            <el-switch v-if="showEnable" @change="enableChanged"
                    v-model="enableSettingData">
            </el-switch>
            <div style="float: right; padding: 3px 0">
                <slot name="header-buttons"></slot>
                <i :class="{'el-icon-arrow-down':!showSetting,'el-icon-arrow-up':showSetting}"
                   @click="showSettingClick"></i>
            </div>
        </div>
        <slot></slot>
    </el-card>
</template>

<script>
    export default {
        name: "SettingCard",
        data() {
            return {
                "showSetting": true,
                "enableSettingData": true,
            }
        },
        created(){
            this.enableSettingData = this.enableSetting;
            this.showSetting = this.enableSetting;
        },
        props: {
            title: {
                type: String,
                default() {
                    return "设置"
                }
            },
            showEnable:{
                type: Boolean,
                default(){
                    return true;
                }
            },
            enableSetting:{
                type: Boolean,
                default(){
                    return true;
                }
            },
        },
        computed: {
            bodyStyle() {
                const display = this.showSetting ? "" : "none";
                return {display};
            }
        },
        watch:{
            enableSetting(val){
                this.enableSettingData = val;
                this.showSetting = val;
            }
        },
        methods:{
            enableChanged(bEnable) {
                this.showSetting = bEnable;
                this.$emit('update:enableSetting', this.enableSettingData);
            },
            showSettingClick() {
                this.showSetting = !this.showSetting;
                if(this.showSetting && !this.enableSettingData){
                    this.enableSettingData = true;
                    this.$emit('update:enableSetting', this.enableSettingData);
                }
            }
        },
    }
</script>

<style>
.setting-card{
    margin-bottom: 10px;
}
</style>
