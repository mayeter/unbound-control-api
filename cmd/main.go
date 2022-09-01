package main

import "github.com/callMe-Root/unbound-control-api/toolbox"

// import "github.com/callMe-Root/unbound-control-api/model"

func main() {

	app := toolbox.App{}
	app.Init()
	app.Run(":8080")

}

// an example usage for arguments
// https://stackoverflow.com/questions/53100425/using-curl-with-commands-in-go

// api/start-updater
// bu endpoint aslında bir bash komutu çalıştırır, yani vereceğimi url e göre farklı görevler alabilir, ama ilk olarak update işlemi için düşünüldü, bu sebeple adı updater
// parametreler: bash komutu
/* updater dosyası için bir url gönderilecek, api bu urlden .sh dosyasını çekip çalıştıracak, bu .sh dosyası şu 3 adımı takip edicek
1- ilgili directory e gidip relases altındaki binary dosyasını curl ile çekicek
2- binary i yerine kopyaladı
3- systemctl restart yapıcak
4- makine adını al, tower apide api/unbound-updated/<makine adı> a gönder
*/

// api/apply-config
// bu aslında bir dosyası yazan endpoint, farklı bir dosyayı farklı bir path e de yazabilir ama ilk olarak config güncellemek için tasarlandığından adı apply-config
// parametreler: dosyanın yazılacağı path, yazılacak dosya (base64)
/* bu endpoint şu adımları takip edicek:
1- base64 stringi byte array e çevirecek
2- path e bu array i file.write edicek
*/

// api/healthcheck
// parametre almayacak
/* yapacağı işlemler:
1- systemctl status unbound
2- systemctl status unbound-control api
3- systemctl status bird
*/

// api/versioncheck
// parametre almayacak
/* yapacağı işlemler:
1- unbound -V | head -n1
2- api versiyonu statik
3- cat /etc/unbound/unbound.conf | head -n1
4- birdc show status | tail -n +2 | head -n1
*/

// api/unbound-control/xxx

/*
Start
Stop
Reload
Stats
dump_cache
Lookup <name>
List_forwards
list_local_zones
list_local_data
view_list_local_zones view
view_list_local_data view
flush <name>
flush_zone <name>
flush_negative     ## bu ekipler yeni dns kaydı yaptık ama çözemiyoruz dedikleri senaryolarda kullanışlı olabilir, her seferinde flush <name> kullanmak yerine ekiplere sunabileceğimiz daha güvenli bir opsiyon olur
forward_add [+i] zone addr ...
forward_remove [+i] zone
local_data name type
local_data_remove name

*/
