<template>
    <span>
        <i class="el-icon-upload2" v-if="!isFirst" @click="moveTop"></i>
        <i class="el-icon-top" v-if="!isFirst" @click="moveUp"></i>
        <i class="el-icon-bottom" v-if="!isLast" @click="moveDown"></i>
        <i class="el-icon-download" v-if="!isLast" @click="moveBottom"></i>
    </span>
</template>

<script>
    export default {
        name: "ItemUpDownIcons",
        computed:{
            isFirst(){
                return this.idx === 0;
            },
            isLast(){
                return this.idx === this.items.length-1;
            }
        },
        methods:{
            moveUp(){
                let delItems = this.items.splice(this.idx, 1);
                this.items.splice(this.idx-1, 0, ...delItems);
            },
            moveTop(){
                let delItems = this.items.splice(this.idx, 1);
                this.items.splice(0, 0, ...delItems);
            },
            moveDown(){
                let delItems = this.items.splice(this.idx, 1);
                this.items.splice(this.idx+1, 0, ...delItems);
            },
            moveBottom(){
                let delItems = this.items.splice(this.idx, 1);
                this.items.push(...delItems);
            },
        },
        props:{
            items:{
                type:Array,
                default(){
                    return [];
                }
            },
            idx:{
                type: Number,
                default(){
                    return 0;
                }
            }
        }
    }
</script>

<style scoped>

</style>
