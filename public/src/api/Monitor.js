import ajax from '@/api/lib/ajax'

export default {
    async listCounters() {
        let res = {
            bSuccess: true,
            msg:""
        };
        let httpRes = await ajax.get("/api/stats");
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
