import ajax from '@/api/lib/ajax'

export default {
    async getPacContent() {
        let res = {
            bSuccess: true,
            msg:""
        };
        let httpRes = await ajax.get("/api/pac");
        if(httpRes.status!==200) {
            res.bSuccess = false;
            res.msg = "http调用失败:"+httpRes.statusText;
        }else{
            res.bSuccess = true;
            res.data = httpRes.data;
        }
        return res;
    },
    async savePac(config) {
        let res = {
            bSuccess: true,
            msg:""
        };
        let httpRes = await ajax.post("/api/pac/save", config);
        if(httpRes.status!==200) {
            res.bSuccess = false;
            res.msg = "http调用失败:"+httpRes.statusText;
        }else{
            res.bSuccess = true;
            res.data = httpRes.data;
        }
        return res;
    },
    async downloadGfwList(config) {
        let res = {
            bSuccess: true,
            msg:""
        };
        let httpRes = await ajax.post("/api/pac/gfwlist/download", config);
        if(httpRes.status!==200) {
            res.bSuccess = false;
            res.msg = "http调用失败:"+httpRes.statusText;
        }else{
            res.bSuccess = true;
            res.data = httpRes.data;
        }
        return res;
    },
    async downloadGeoDat(config) {
        let res = {
            bSuccess: true,
            msg:""
        };
        let httpRes = await ajax.post("/api/pac/geodat/download", config);
        if(httpRes.status!==200) {
            res.bSuccess = false;
            res.msg = "http调用失败:"+httpRes.statusText;
        }else{
            res.bSuccess = true;
            res.data = httpRes.data;
        }
        return res;
    },
    async loadConfig() {
        let res = {
            bSuccess: true,
            msg:""
        };
        let httpRes = await ajax.get("/api/pac/config");
        if(httpRes.status!==200) {
            res.bSuccess = false;
            res.msg = "http调用失败:"+httpRes.statusText;
        }else{
            res.bSuccess = true;
            res.data = httpRes.data;
        }
        return res;
    },

}
