import ajax from '@/api/lib/ajax'

export default {
    async getConfig() {
        let res = {
            bSuccess: true,
            msg:""
        };
        let httpRes = await ajax.get("/api/server/config");
        if(httpRes.status!==200) {
            res.bSuccess = false;
            res.msg = "http调用失败:"+httpRes.statusText;
        }else{
            res.bSuccess = true;
            res.data = httpRes.data;
        }
        return res;
    },
    async updateConfig(data) {
        let res = {
            bSuccess: true,
            msg:""
        };
        let httpRes = await ajax.post("/api/server/config",JSON.stringify(data, null, 2))
        if(httpRes.status!==200) {
            res.bSuccess = false;
            res.msg = "http调用失败:"+httpRes.statusText;
        }else{
            res.bSuccess = true;
        }
        return res;
    },
    async reloadServer() {
        let res = {
            bSuccess: true,
            msg:""
        };
        let httpRes = await ajax.get("/api/server/reload");
        if(httpRes.status!==200) {
            res.bSuccess = false;
            res.msg = "http调用失败:"+httpRes.statusText;
        }else{
            res.bSuccess = true;
        }
        return res;
    },

}
