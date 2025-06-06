package handler

import (
	"fmt"
	"minik8s/pkg/etcd"
	nginxmanager "minik8s/pkg/nginx/app"

	"encoding/json"
	"minik8s/pkg/apiobj"
	"minik8s/pkg/message"

	"github.com/gin-gonic/gin"
)

func GetGlobalDns(c *gin.Context) {
	fmt.Println("getGlobalDns")
	key := etcd.PATH_EtcdDns
	var resList []string
	resList, err := etcd.EtcdKV.GetPrefix(key)
	if err != nil {
		c.JSON(500, gin.H{"get": "fail"})
	}
	c.JSON(200, gin.H{
		"data": resList,
	})

}

func GetAllDns(c *gin.Context) {
	fmt.Println("getAllDns")
	namespace := c.Param("namespace")
	key := fmt.Sprintf(etcd.PATH_EtcdDns+"/%s", namespace)
	var resList []string
	resList, err := etcd.EtcdKV.GetPrefix(key)
	if err != nil {
		c.JSON(500, gin.H{"get": "fail"})
	}
	c.JSON(200, gin.H{
		"data": resList,
	})
}

func AddDns(c *gin.Context) {
	fmt.Println("addDns")
	var dns apiobj.Dns
	c.ShouldBind(&dns)
	namespace := c.Param("namespace")
	name := c.Param("name")
	key := fmt.Sprintf(etcd.PATH_EtcdDns+"/%s/%s", namespace, name)

	for id, path := range dns.Spec.Paths {
		svc_name := path.ServiceName
		svc_key := fmt.Sprintf(etcd.PATH_EtcdServices+"/%s/%s", namespace, svc_name)

		var svc apiobj.Service
		svcJson, _ := etcd.EtcdKV.Get(svc_key)
		if svcJson == nil {
			c.JSON(500, gin.H{"svc": "not found"})
			return
		}
		json.Unmarshal(svcJson, &svc)

		fmt.Println(svc.Spec.ClusterIP)
		path.ServiceIp = svc.Spec.ClusterIP
		dns.Spec.Paths[id] = path
	}

	// add server block
	nginxmanager.AddServerBlock(dns.Spec.Host, dns.Spec.Paths)

	dnsJson, err := json.Marshal(dns)
	if err != nil {
		c.JSON(500, gin.H{"add": "fail"})
	}
	etcd.EtcdKV.Put(key, dnsJson)
	c.JSON(200, gin.H{"add": string(dnsJson)})

	res, err := etcd.EtcdKV.Get(etcd.PATH_EtcdDnsNginxIP)
	if err != nil {
		fmt.Println("get etcd error")
	}
	nginxIp := string(res)
	msg := message.Message{
		Type:    "Add",
		URL:     key,
		Name:    dns.Spec.Host,
		Content: nginxIp,
	}

	msgJson, _ := json.Marshal(msg)
	p := message.NewPublisher()
	defer p.Close()

	nodeKey := etcd.PATH_EtcdNodes
	resList, _ := etcd.EtcdKV.GetPrefix(nodeKey)

	for _, item := range resList {
		var node apiobj.Node
		json.Unmarshal([]byte(item), &node)
		que := fmt.Sprintf(message.DnsQueue+"-%s", node.MetaData.Name)
		p.Publish(que, msgJson)
	}
}

func DeleteDns(c *gin.Context) {
	fmt.Println("deleteDns")
	namespace := c.Param("namespace")
	name := c.Param("name")
	key := fmt.Sprintf(etcd.PATH_EtcdDns+"/%s/%s", namespace, name)
	// delete server block
	var dns apiobj.Dns
	dnsJson, err := etcd.EtcdKV.Get(key)
	if err != nil {
		c.JSON(500, gin.H{"delete": "fail"})
	}
	err = json.Unmarshal(dnsJson, &dns)
	if err != nil {
		c.JSON(500, gin.H{"delete": "fail"})
	}
	nginxmanager.DeleteServerBlock(dns.Spec.Host)

	err = etcd.EtcdKV.Delete(key)
	if err != nil {
		c.JSON(500, gin.H{"delete": "fail"})
	}
	c.JSON(200, gin.H{"delete": "success"})

	msg := message.Message{
		Type:    "Delete",
		URL:     key,
		Name:    dns.Spec.Host,
		Content: "",
	}
	msgJson, _ := json.Marshal(msg)
	p := message.NewPublisher()
	defer p.Close()

	nodeKey := etcd.PATH_EtcdNodes
	resList, _ := etcd.EtcdKV.GetPrefix(nodeKey)

	for _, item := range resList {
		var node apiobj.Node
		json.Unmarshal([]byte(item), &node)
		que := fmt.Sprintf(message.DnsQueue+"-%s", node.MetaData.Name)
		p.Publish(que, msgJson)
	}
}

func UpdateDns(c *gin.Context) {
	fmt.Println("updateDns")
	var dns apiobj.Dns
	c.ShouldBind(&dns)
	namespace := c.Param("namespace")
	name := c.Param("name")
	key := fmt.Sprintf(etcd.PATH_EtcdDns+"/%s/%s", namespace, name)

	dnsJson, err := json.Marshal(dns)
	if err != nil {
		c.JSON(500, gin.H{"update": "fail"})
	}
	err = etcd.EtcdKV.Put(key, dnsJson)
	if err != nil {
		c.JSON(500, gin.H{"update": "fail"})
	}
	c.JSON(200, gin.H{"update": "success"})
}

func GetDns(c *gin.Context) {
	fmt.Print("getDns")
	namespace := c.Param("namespace")
	name := c.Param("name")
	key := fmt.Sprintf(etcd.PATH_EtcdDns+"/%s/%s", namespace, name)
	res, err := etcd.EtcdKV.Get(key)
	if err != nil {
		c.JSON(500, gin.H{"get": "fail"})
	}
	c.JSON(200, gin.H{"data": string(res)})
}

func GetDnsStatus(c *gin.Context) {
	fmt.Println("getDnsStatus")
	namespace := c.Param("namespace")
	name := c.Param("name")
	key := fmt.Sprintf(etcd.PATH_EtcdDns+"/%s/%s/status", namespace, name)
	res, err := etcd.EtcdKV.Get(key)
	if err != nil {
		c.JSON(500, gin.H{"get": "fail"})
	}
	c.JSON(200, gin.H{"data": string(res)})
}

func UpdateDnsStatus(c *gin.Context) {
	fmt.Println("updateDnsStatus")

	var dnsStatus apiobj.DnsStatus
	c.ShouldBind(&dnsStatus)
	namespace := c.Param("namespace")
	name := c.Param("name")

	key := fmt.Sprintf(etcd.PATH_EtcdDns+"/%s/%s", namespace, name)
	res, err := etcd.EtcdKV.Get(key)
	if err != nil {
		c.JSON(500, gin.H{"get": "fail"})
	}
	var dns apiobj.Dns
	json.Unmarshal([]byte(res), &dns)
	dns.Status = dnsStatus

	dnsJson, _ := json.Marshal(dns)
	etcd.EtcdKV.Put(key, dnsJson)
	c.JSON(200, gin.H{"update": string(dnsJson)})
}
