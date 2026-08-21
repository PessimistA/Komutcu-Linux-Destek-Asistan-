// komutcu - CachyOS / Arch Linux komut asistanı
// Tamamen offline çalışır, hiçbir bağımlılığı yoktur.
package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strings"
	"time"
)

// ---------- Veri modeli ----------

type Cmd struct {
	Kat    string   // kategori
	Is     string   // ne işe yarar (Türkçe)
	Komut  string   // komutun kendisi
	Not    string   // açıklama / ipucu
	Anahtar []string // arama anahtar kelimeleri
	Risk   int      // 0 = güvenli, 1 = dikkat, 2 = tehlikeli
}

const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	dim    = "\033[2m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	blue   = "\033[34m"
	cyan   = "\033[36m"
)

var kategoriler = map[string]string{
	"paket":     "Paket yönetimi (pacman, paru, flatpak)",
	"sistem":    "Sistem yönetimi (systemd, kernel, güncelleme)",
	"dosya":     "Dosya ve dizin işlemleri",
	"disk":      "Disk, bölüm, btrfs ve snapshot",
	"ag":        "Ağ, wifi, ssh, indirme",
	"surec":     "Süreç (process) yönetimi",
	"kullanici": "Kullanıcı, izin, sudo",
	"log":       "Log ve hata ayıklama",
	"donanim":   "Donanım, sürücü, GPU",
	"cachyos":   "CachyOS'a özel araçlar",
	"oyun":      "Oyun, Steam, Proton",
	"bakim":     "Temizlik ve bakım",
	"metin":     "Metin işleme ve editörler",
	"kabuk":     "Kabuk (shell) hileleri",
}

// ---------- Komut veritabanı ----------

var db = []Cmd{
	// ---------------- PAKET ----------------
	{"paket", "Sistemi tamamen güncelle", "sudo pacman -Syu", "Arch'ta güncellemenin TEK doğru yolu. Asla sadece -Sy kullanma, sistemi kırar.", []string{"güncelle", "update", "upgrade", "yükselt", "yeni sürüm"}, 0},
	{"paket", "Paket kur", "sudo pacman -S paket_adi", "Birden fazla paketi boşlukla ayırarak yazabilirsin: sudo pacman -S firefox vlc", []string{"kur", "yükle", "install", "indir", "program"}, 0},
	{"paket", "Paket kaldır (ayarları ve gereksiz bağımlılıkları ile)", "sudo pacman -Rns paket_adi", "-R sadece paketi siler, -Rns bağımlılıkları ve config dosyalarını da temizler.", []string{"sil", "kaldır", "remove", "uninstall", "temizle"}, 1},
	{"paket", "Paket ara (depoda)", "pacman -Ss arama_kelimesi", "Küçük s ile arama yapılır. Örn: pacman -Ss video editor", []string{"ara", "bul", "search", "hangi paket"}, 0},
	{"paket", "Kurulu paketleri ara", "pacman -Qs arama_kelimesi", "Q = query, sistemde kurulu olanlarda arar.", []string{"kurulu", "yüklü", "listele", "hangi programlar"}, 0},
	{"paket", "Bir paket hakkında bilgi al", "pacman -Si paket_adi", "Kurulu bir paket için: pacman -Qi paket_adi", []string{"bilgi", "detay", "info", "nedir", "boyut"}, 0},
	{"paket", "Bir dosyanın hangi pakete ait olduğunu bul", "pacman -Qo /yol/dosya", "Örn: pacman -Qo /usr/bin/ffmpeg", []string{"hangi paket", "sahip", "ait", "dosya nereden"}, 0},
	{"paket", "AUR'dan paket kur (paru ile)", "paru -S paket_adi", "CachyOS'ta paru hazır gelir. AUR paketleri topluluk tarafından yazılır, PKGBUILD'i okumadan kurmamaya çalış.", []string{"aur", "paru", "yay", "topluluk", "kur"}, 1},
	{"paket", "AUR dahil her şeyi güncelle", "paru -Syu", "pacman -Syu + AUR paketleri birlikte güncellenir.", []string{"aur güncelle", "paru güncelle", "hepsini güncelle"}, 0},
	{"paket", "Flatpak uygulaması kur", "flatpak install flathub uygulama.adi", "Örn: flatpak install flathub com.spotify.Client", []string{"flatpak", "sandbox", "flathub", "kur"}, 0},
	{"paket", "Flatpak uygulamalarını güncelle", "flatpak update", "Flatpak, pacman'dan bağımsızdır; ayrıca güncellenmesi gerekir.", []string{"flatpak güncelle", "update"}, 0},
	{"paket", "Bir paketin hangi dosyaları kurduğunu gör", "pacman -Ql paket_adi", "Programın çalıştırılabilir dosyasını bulmak için: pacman -Ql paket | grep bin/", []string{"dosyalar", "içerik", "nereye kuruldu", "yol"}, 0},
	{"paket", "Paket veritabanı kilidini kaldır", "sudo rm /var/lib/pacman/db.lck", "'unable to lock database' hatası alırsan kullan. Önce başka bir pacman çalışmadığından emin ol.", []string{"kilit", "lock", "db.lck", "hata", "unable to lock"}, 1},
	{"paket", "Anahtarlık (keyring) sorununu düzelt", "sudo pacman -S archlinux-keyring cachyos-keyring && sudo pacman-key --populate", "İmza doğrulama hatalarının (invalid or corrupted package) klasik çözümü.", []string{"keyring", "imza", "signature", "anahtar", "corrupted", "pgp"}, 1},
	{"paket", "Belirli bir paketi güncellemeden hariç tut", "sudo nano /etc/pacman.conf", "IgnorePkg = paket_adi satırını ekle. Uzun süre kullanma, kısmi güncelleme riskli.", []string{"hariç", "ignore", "güncelleme", "dondur", "sabitle"}, 1},

	// ---------------- SISTEM ----------------
	{"sistem", "Servisi başlat", "sudo systemctl start servis", "Örn: sudo systemctl start bluetooth", []string{"servis", "başlat", "start", "service", "daemon"}, 0},
	{"sistem", "Servisi durdur", "sudo systemctl stop servis", "Anlık durdurur, açılışta tekrar başlar.", []string{"durdur", "stop", "kapat", "servis"}, 0},
	{"sistem", "Servisi açılışta otomatik başlat", "sudo systemctl enable --now servis", "--now hem şimdi başlatır hem açılışa ekler.", []string{"otomatik", "enable", "açılışta", "boot", "başlangıç"}, 0},
	{"sistem", "Servisi açılıştan kaldır", "sudo systemctl disable --now servis", "Gereksiz servisleri kapatmak açılışı hızlandırır.", []string{"disable", "kapat", "otomatik başlama"}, 0},
	{"sistem", "Servisin durumunu gör", "systemctl status servis", "Çalışıyor mu, hata var mı, son log satırları burada görünür.", []string{"durum", "status", "çalışıyor mu", "servis"}, 0},
	{"sistem", "Çalışan tüm servisleri listele", "systemctl list-units --type=service --state=running", "Sistemde ne varsa görürsün.", []string{"servisler", "liste", "çalışan", "running"}, 0},
	{"sistem", "Hata veren servisleri gör", "systemctl --failed", "Bir şey çalışmıyorsa ilk bakılacak yer burası.", []string{"hata", "failed", "bozuk", "çalışmıyor"}, 0},
	{"sistem", "Sistemi yeniden başlat / kapat", "systemctl reboot   |   systemctl poweroff", "sudo gerekmez, normal kullanıcı yapabilir.", []string{"reboot", "restart", "kapat", "yeniden başlat", "shutdown"}, 0},
	{"sistem", "Açılış süresini analiz et", "systemd-analyze blame", "Hangi servis açılışı yavaşlatıyor görürsün.", []string{"açılış", "yavaş", "boot", "hız", "süre"}, 0},
	{"sistem", "Kurulu kernel'ları gör", "pacman -Q | grep linux", "CachyOS'ta linux-cachyos, linux-cachyos-lts gibi çekirdekler olur.", []string{"kernel", "çekirdek", "linux", "sürüm"}, 0},
	{"sistem", "Çalışan kernel sürümünü öğren", "uname -r", "Sistem bilgisinin tamamı için: uname -a", []string{"kernel sürümü", "uname", "versiyon", "hangi çekirdek"}, 0},
	{"sistem", "Sistem bilgisi özeti", "fastfetch", "CachyOS'ta hazır gelir. Ekran görüntüsü paylaşırken kullanılan şey budur.", []string{"sistem bilgisi", "fastfetch", "neofetch", "özet", "donanım"}, 0},
	{"sistem", "Saat ve zaman dilimini ayarla", "sudo timedatectl set-timezone Europe/Istanbul", "Windows ile çift boot yapıyorsan: sudo timedatectl set-local-rtc 1", []string{"saat", "zaman", "timezone", "tarih", "istanbul"}, 0},
	{"sistem", "Sistem dilini/locale ayarla", "sudo nano /etc/locale.gen && sudo locale-gen", "tr_TR.UTF-8 satırının başındaki # işaretini kaldır, sonra locale-gen çalıştır.", []string{"dil", "locale", "türkçe", "karakter"}, 1},
	{"sistem", "Bootloader yapılandırmasını yenile (GRUB)", "sudo grub-mkconfig -o /boot/grub/grub.cfg", "CachyOS'ta varsayılan Limine ise buna gerek yok; limine kendi senkronunu yapar.", []string{"grub", "bootloader", "önyükleyici", "boot menü"}, 1},

	// ---------------- DOSYA ----------------
	{"dosya", "Dizin içeriğini listele", "ls -lah", "-l detay, -a gizli dosyalar, -h okunabilir boyut. En çok kullanacağın kombinasyon.", []string{"listele", "ls", "dosyalar", "içerik", "dizin"}, 0},
	{"dosya", "Dizin değiştir", "cd /yol/dizin", "cd ~ ev dizinine, cd - bir öncekine, cd .. bir üste gider.", []string{"cd", "git", "dizin", "klasör", "geç"}, 0},
	{"dosya", "Bulunduğun dizini göster", "pwd", "Nerede olduğunu kaybettiğinde.", []string{"pwd", "neredeyim", "konum", "yol"}, 0},
	{"dosya", "Klasör oluştur (iç içe dahil)", "mkdir -p /yol/a/b/c", "-p ara klasörleri de oluşturur.", []string{"klasör", "oluştur", "mkdir", "yeni dizin"}, 0},
	{"dosya", "Dosya kopyala / klasör kopyala", "cp dosya hedef   |   cp -r klasor hedef", "-r klasörler için zorunlu, -v ne yaptığını gösterir.", []string{"kopyala", "cp", "copy", "çoğalt"}, 0},
	{"dosya", "Taşı veya yeniden adlandır", "mv kaynak hedef", "Aynı komut hem taşır hem isim değiştirir.", []string{"taşı", "mv", "yeniden adlandır", "isim değiştir", "rename"}, 1},
	{"dosya", "Dosya/klasör sil", "rm dosya   |   rm -r klasor", "Linux'ta çöp kutusu YOK. Silmeden önce iki kez düşün. rm -rf / gibi şeyleri asla yazma.", []string{"sil", "rm", "delete", "kaldır"}, 2},
	{"dosya", "Dosya ara (isme göre)", "find /yol -name '*.pdf'", "Ev dizininde ara: find ~ -iname '*rapor*' (-iname büyük/küçük harf duyarsız)", []string{"ara", "bul", "find", "dosya nerede", "arama"}, 0},
	{"dosya", "Hızlı dosya arama (veritabanı ile)", "locate dosya_adi", "İlk kullanımdan önce: sudo pacman -S plocate && sudo updatedb", []string{"locate", "hızlı ara", "bul"}, 0},
	{"dosya", "Dosya içeriğini oku", "cat dosya   |   less dosya", "Uzun dosyalar için less kullan, q ile çıkarsın.", []string{"oku", "göster", "cat", "less", "içerik"}, 0},
	{"dosya", "Arşiv oluştur / aç (tar.gz)", "tar -czf ars.tar.gz klasor   |   tar -xzf ars.tar.gz", "c=create, x=extract, z=gzip, f=file. zip için: unzip dosya.zip", []string{"arşiv", "tar", "sıkıştır", "zip", "aç", "extract"}, 0},
	{"dosya", "Dosya boyutlarını gör", "du -sh *", "Bulunduğun dizindeki her şeyin boyutu. Büyükten küçüğe: du -sh * | sort -rh", []string{"boyut", "yer kaplıyor", "du", "büyük dosya", "doluluk"}, 0},
	{"dosya", "Sembolik link oluştur", "ln -s /gercek/yol /kisayol", "Kısayol gibi düşün, -s olmadan hard link olur.", []string{"link", "kısayol", "symlink", "bağlantı"}, 0},
	{"dosya", "Yedekleme / senkronizasyon", "rsync -avh --progress kaynak/ hedef/", "Kaynağın sonundaki / önemlidir: içeriğini kopyalar, klasörün kendisini değil.", []string{"yedek", "rsync", "senkron", "backup", "kopyala"}, 1},

	// ---------------- DISK ----------------
	{"disk", "Diskleri ve bölümleri listele", "lsblk -f", "-f dosya sistemi ve UUID'leri de gösterir.", []string{"disk", "bölüm", "partition", "lsblk", "usb"}, 0},
	{"disk", "Boş disk alanını gör", "df -h", "-h insan okuyabilir birimlerde (GB) gösterir.", []string{"boş alan", "doluluk", "df", "disk dolu", "kaç gb"}, 0},
	{"disk", "USB / disk bağla (mount)", "sudo mount /dev/sdb1 /mnt", "Çıkarmak için: sudo umount /mnt", []string{"mount", "bağla", "usb", "takma", "harici disk"}, 1},
	{"disk", "Otomatik bağlanacak diskleri ayarla", "sudo nano /etc/fstab", "Hatalı fstab sistemi açılmaz hale getirir. Düzenlemeden önce yedek al: sudo cp /etc/fstab /etc/fstab.bak", []string{"fstab", "otomatik bağla", "kalıcı mount"}, 2},
	{"disk", "Btrfs alt birimlerini (subvolume) listele", "sudo btrfs subvolume list /", "CachyOS varsayılan olarak btrfs kullanır.", []string{"btrfs", "subvolume", "alt birim"}, 0},
	{"disk", "Btrfs gerçek disk kullanımını gör", "sudo btrfs filesystem usage /", "df btrfs'te yanıltıcı olabilir, doğru rakam burada.", []string{"btrfs", "kullanım", "boş alan", "gerçek boyut"}, 0},
	{"disk", "Snapshot (anlık görüntü) listele", "sudo snapper -c root list", "CachyOS'ta snapper kuruluysa güncelleme öncesi otomatik snapshot alınır.", []string{"snapshot", "snapper", "yedek", "geri dön", "anlık görüntü"}, 0},
	{"disk", "Elle snapshot al", "sudo snapper -c root create -d 'aciklama'", "Riskli bir işlem yapmadan önce mutlaka al.", []string{"snapshot al", "yedek al", "snapper create"}, 0},
	{"disk", "Snapshot'ları boot menüsüne yaz (Limine)", "sudo limine-snapper-sync", "CachyOS'un Limine bootloader'ında snapshot'tan açılış için gerekir.", []string{"limine", "snapshot boot", "geri yükleme", "sync"}, 0},
	{"disk", "Disk sağlığını kontrol et (SMART)", "sudo smartctl -a /dev/nvme0n1", "Önce: sudo pacman -S smartmontools", []string{"smart", "disk sağlığı", "arıza", "ssd ömrü"}, 0},
	{"disk", "En çok yer kaplayanları görsel bul", "ncdu /", "Önce: sudo pacman -S ncdu. Ok tuşlarıyla gezinirsin, disk dolduğunda hayat kurtarır.", []string{"ncdu", "disk dolu", "büyük klasör", "yer kaplıyor"}, 0},

	// ---------------- AG ----------------
	{"ag", "IP adresini öğren", "ip a", "Yerel IP burada. Dış IP için: curl ifconfig.me", []string{"ip", "adres", "network", "ağ", "ip adresim"}, 0},
	{"ag", "Wifi ağlarını listele ve bağlan", "nmcli device wifi list && nmcli device wifi connect 'AGADI' password 'SIFRE'", "Terminalden wifi yönetimi. Grafik istersen: nmtui", []string{"wifi", "kablosuz", "bağlan", "internet", "nmcli"}, 0},
	{"ag", "Terminal tabanlı ağ arayüzü", "nmtui", "Ok tuşlarıyla wifi seçip bağlanabileceğin basit menü.", []string{"nmtui", "wifi menü", "ağ arayüz"}, 0},
	{"ag", "İnternet bağlantısını test et", "ping -c 4 archlinux.org", "-c 4 dört paket gönderip durur.", []string{"ping", "internet var mı", "test", "bağlantı"}, 0},
	{"ag", "Dosya indir", "curl -LO https://adres/dosya", "Alternatif: wget https://adres/dosya", []string{"indir", "download", "curl", "wget"}, 0},
	{"ag", "Uzak sunucuya bağlan", "ssh kullanici@sunucu_ip", "Port farklıysa: ssh -p 2222 kullanici@ip", []string{"ssh", "uzak", "sunucu", "bağlan", "remote"}, 0},
	{"ag", "SSH ile dosya kopyala", "scp dosya kullanici@ip:/hedef/yol", "Klasör için -r ekle.", []string{"scp", "dosya gönder", "uzak kopyala"}, 0},
	{"ag", "Hangi portlar açık / dinleniyor", "ss -tulpn", "sudo ile çalıştırırsan hangi programın dinlediğini de görürsün.", []string{"port", "dinleme", "ss", "netstat", "açık port"}, 0},
	{"ag", "DNS sorunlarını çöz", "resolvectl status", "DNS'i elle ayarlamak için NetworkManager kullan, /etc/resolv.conf'u elle düzenleme.", []string{"dns", "isim çözümleme", "site açılmıyor", "resolv"}, 1},
	{"ag", "İndirme aynalarını (mirror) hızlandır", "sudo cachyos-rate-mirrors", "CachyOS'a özel; en hızlı sunucuları seçer. Güncelleme yavaşsa ilk deneyeceğin şey.", []string{"mirror", "yavaş indirme", "ayna", "hızlandır", "cachyos"}, 0},

	// ---------------- SUREC ----------------
	{"surec", "Çalışan süreçleri canlı izle", "htop", "F9 ile süreç öldürme, F6 ile sıralama. btop de kuruluysa daha şık.", []string{"htop", "top", "işlemci", "ram", "süreç", "görev yöneticisi"}, 0},
	{"surec", "Bir programı adıyla öldür", "pkill -f program_adi", "Sertçe kapatmak için: pkill -9 -f program_adi", []string{"kapat", "öldür", "kill", "donmuş", "yanıt vermiyor"}, 1},
	{"surec", "Bir sürecin PID'sini bul", "pgrep -a firefox", "-a komutun tamamını gösterir.", []string{"pid", "süreç numarası", "pgrep", "bul"}, 0},
	{"surec", "PID ile süreç öldür", "kill -9 PID", "-9 sinyali kaydetmeden kapatır, son çare olarak kullan.", []string{"kill", "pid", "öldür", "zorla kapat"}, 1},
	{"surec", "Kim CPU/RAM yiyor?", "ps aux --sort=-%mem | head", "-%cpu yazarsan işlemciye göre sıralar.", []string{"ram", "cpu", "yavaş", "kim yiyor", "bellek"}, 0},
	{"surec", "Programı arka planda çalıştır", "komut &   veya   nohup komut &", "nohup terminali kapatsan da devam etmesini sağlar.", []string{"arka plan", "background", "nohup", "devam"}, 0},
	{"surec", "Donmuş masaüstünü yeniden başlat (Wayland/KDE)", "systemctl --user restart plasma-plasmashell", "Oturumu kapatmadan panelin/kabuğun düzelmesini sağlar.", []string{"donma", "plasma", "kde", "masaüstü", "panel gitti"}, 1},

	// ---------------- KULLANICI ----------------
	{"kullanici", "Root olarak komut çalıştır", "sudo komut", "Bir önceki komutu sudo ile tekrarla: sudo !!", []string{"sudo", "root", "yetki", "izin", "yönetici"}, 1},
	{"kullanici", "Dosya izinlerini değiştir", "chmod 755 dosya", "755 = sahip her şeyi, diğerleri okuma+çalıştırma. Betiği çalıştırılabilir yap: chmod +x betik.sh", []string{"izin", "chmod", "permission", "çalıştırılabilir", "denied"}, 1},
	{"kullanici", "Dosya sahibini değiştir", "sudo chown kullanici:grup dosya", "Klasörün tamamı için -R ekle.", []string{"sahip", "chown", "owner", "kullanıcı değiştir"}, 1},
	{"kullanici", "Kendi kullanıcı adını ve gruplarını gör", "id", "Hangi gruplarda olduğun burada. Örn: wheel, video, input", []string{"kullanıcı", "grup", "id", "kimim"}, 0},
	{"kullanici", "Kullanıcıyı bir gruba ekle", "sudo usermod -aG grup kullanici", "-a olmadan diğer gruplardan atılırsın! Değişiklik için yeniden giriş yap.", []string{"grup ekle", "usermod", "yetki ver", "docker grubu"}, 1},
	{"kullanici", "Şifre değiştir", "passwd", "Başka kullanıcı için: sudo passwd kullanici", []string{"şifre", "parola", "password", "değiştir"}, 0},

	// ---------------- LOG ----------------
	{"log", "Son sistem loglarını canlı izle", "journalctl -f", "Bir şey denerken açık bırak, hatayı anında görürsün. Ctrl+C ile çık.", []string{"log", "journalctl", "hata", "izle", "canlı"}, 0},
	{"log", "Bu açılışın loglarını gör", "journalctl -b", "Bir önceki açılış için: journalctl -b -1", []string{"boot log", "açılış hatası", "journalctl"}, 0},
	{"log", "Sadece hataları göster", "journalctl -p 3 -b", "-p 3 = error seviyesi ve üstü. Sorun ararken en verimli komut.", []string{"hata", "error", "sorun", "log", "kritik"}, 0},
	{"log", "Bir servisin loglarını gör", "journalctl -u servis -e", "-e sonuna atlar. Örn: journalctl -u bluetooth -e", []string{"servis log", "journalctl", "hata", "unit"}, 0},
	{"log", "Kernel/donanım mesajları", "sudo dmesg -w", "USB takıp çıkarınca ne oluyor görmek için ideal.", []string{"dmesg", "kernel", "donanım", "usb", "mesaj"}, 0},
	{"log", "Log boyutunu küçült", "sudo journalctl --vacuum-size=200M", "Loglar GB'lara çıkabilir; disk temizlerken işe yarar.", []string{"log temizle", "vacuum", "yer aç", "journal büyük"}, 0},

	// ---------------- DONANIM ----------------
	{"donanim", "Donanım sürücülerini otomatik kur (CachyOS)", "sudo chwd -a", "chwd, CachyOS'un sürücü aracıdır (Manjaro'daki mhwd'nin karşılığı).", []string{"sürücü", "driver", "chwd", "donanım", "otomatik kur"}, 1},
	{"donanim", "Ekran kartını öğren", "lspci -k | grep -A 3 -i vga", "Hangi sürücünün (kernel driver in use) yüklü olduğunu da gösterir.", []string{"ekran kartı", "gpu", "nvidia", "amd", "vga"}, 0},
	{"donanim", "Tüm PCI/USB cihazları listele", "lspci   |   lsusb", "Cihaz tanınmıyorsa buradan başlanır.", []string{"cihaz", "lspci", "lsusb", "donanım listesi"}, 0},
	{"donanim", "CPU bilgisi", "lscpu", "Çekirdek sayısı, mimari, frekans.", []string{"işlemci", "cpu", "çekirdek", "lscpu"}, 0},
	{"donanim", "RAM durumu", "free -h", "Btrfs/zram kullanımını da göz önünde bulundur.", []string{"ram", "bellek", "free", "memory"}, 0},
	{"donanim", "Sıcaklıkları gör", "sensors", "Önce: sudo pacman -S lm_sensors && sudo sensors-detect", []string{"sıcaklık", "ısı", "fan", "sensors", "termal"}, 0},
	{"donanim", "NVIDIA durumunu kontrol et", "nvidia-smi", "Sürücü düzgün yüklendiyse tablo çıkar; komut bulunamazsa sürücü yok demektir.", []string{"nvidia", "gpu", "sürücü", "ekran kartı"}, 0},
	{"donanim", "Ses aygıtlarını yönet (PipeWire)", "wpctl status", "Varsayılan çıkışı değiştirmek için: wpctl set-default ID", []string{"ses", "audio", "pipewire", "hoparlör", "mikrofon"}, 0},
	{"donanim", "Bluetooth cihazı eşleştir", "bluetoothctl", "İçeride: scan on → pair MAC → trust MAC → connect MAC", []string{"bluetooth", "kulaklık", "eşleştir", "bağlan"}, 0},

	// ---------------- CACHYOS ----------------
	{"cachyos", "CachyOS karşılama/kurulum merkezini aç", "cachyos-hello", "Sürücü kurulumu, paket önerileri ve sistem ayarları tek yerde.", []string{"cachyos", "hello", "başlangıç", "kurulum", "merkez"}, 0},
	{"cachyos", "Kernel yöneticisini aç", "cachyos-kernel-manager", "Farklı çekirdekleri (LTS, RT, BORE vb.) grafik arayüzle kurup kaldırırsın.", []string{"kernel", "çekirdek", "değiştir", "kernel manager", "lts"}, 1},
	{"cachyos", "Aynaları en hızlıya ayarla", "sudo cachyos-rate-mirrors", "Hem CachyOS hem Arch aynalarını sıralar. Ayda bir çalıştırmak iyi fikir.", []string{"mirror", "hız", "yavaş", "ayna", "indirme"}, 0},
	{"cachyos", "CachyOS depolarının aktif olduğunu doğrula", "grep -A1 'cachyos' /etc/pacman.conf", "Depo sırası önemlidir; cachyos depoları core'dan önce gelmelidir.", []string{"depo", "repo", "pacman.conf", "cachyos"}, 0},
	{"cachyos", "Sistem optimizasyon ayarlarını gör (ananicy)", "systemctl status ananicy-cpp", "CachyOS süreç önceliklerini otomatik ayarlar, masaüstü akıcılığını sağlayan şey budur.", []string{"ananicy", "optimizasyon", "performans", "öncelik"}, 0},
	{"cachyos", "CPU güç profilini değiştir", "powerprofilesctl set performance", "Seçenekler: performance, balanced, power-saver. Listelemek için: powerprofilesctl list", []string{"performans", "güç", "batarya", "profil", "power"}, 0},
	{"cachyos", "CachyOS sürüm ve depo bilgisi", "cat /etc/os-release", "Sistem gerçekten CachyOS mu, hangi sürümde görürsün.", []string{"sürüm", "versiyon", "os-release", "cachyos"}, 0},

	// ---------------- OYUN ----------------
	{"oyun", "Steam kur", "sudo pacman -S steam", "Multilib deposunun açık olması gerekir; CachyOS'ta varsayılan olarak açıktır.", []string{"steam", "oyun", "kur", "gaming"}, 0},
	{"oyun", "Oyun paketlerini toplu kur", "sudo pacman -S cachyos-gaming-meta", "Wine, Proton, Lutris, MangoHud ve gerekli kütüphaneleri tek seferde getirir.", []string{"oyun", "gaming", "meta", "wine", "lutris", "toplu"}, 0},
	{"oyun", "Proton-CachyOS kur", "sudo pacman -S proton-cachyos", "Steam'de oyun ayarlarından Compatibility kısmında seçilebilir hale gelir.", []string{"proton", "wine", "uyumluluk", "oyun çalışmıyor"}, 0},
	{"oyun", "Oyunda FPS/sıcaklık göster", "mangohud %command%", "Bu satırı Steam'de oyunun Launch Options kısmına yaz.", []string{"fps", "mangohud", "sayaç", "performans", "overlay"}, 0},
	{"oyun", "Oyunu gamemode ile başlat", "gamemoderun %command%", "MangoHud ile birlikte: gamemoderun mangohud %command%", []string{"gamemode", "performans", "oyun hızlandır"}, 0},
	{"oyun", "Vulkan sürücüsünü doğrula", "vulkaninfo | grep deviceName", "Önce: sudo pacman -S vulkan-tools. Boş çıkarsa GPU sürücüsü eksiktir.", []string{"vulkan", "gpu", "oyun açılmıyor", "sürücü test"}, 0},

	// ---------------- BAKIM ----------------
	{"bakim", "Yetim (orphan) paketleri temizle", "sudo pacman -Rns $(pacman -Qtdq)", "Hiçbir şeyin kullanmadığı bağımlılıkları siler. Çıktı boşsa hata verir, normaldir.", []string{"temizlik", "orphan", "yetim", "gereksiz", "yer aç"}, 1},
	{"bakim", "Paket önbelleğini temizle", "sudo paccache -rk2", "Her paketin son 2 sürümünü tutar, gerisini siler. Genelde GB'larca yer açar.", []string{"cache", "önbellek", "temizle", "yer aç", "paccache"}, 0},
	{"bakim", "Kullanılmayan flatpak verilerini sil", "flatpak uninstall --unused", "Eski runtime'lar ciddi yer kaplar.", []string{"flatpak", "temizle", "unused", "yer"}, 0},
	{"bakim", "Yapılandırma dosyası farklarını kontrol et", "sudo pacdiff", "Güncelleme sonrası .pacnew dosyalarını birleştirmek için. DIFFPROG=nvim sudo -E pacdiff daha kullanışlı.", []string{"pacnew", "pacdiff", "config", "güncelleme sonrası"}, 1},
	{"bakim", "Arch haber/uyarılarını kontrol et", "curl -s https://archlinux.org/feeds/news/ | head -40", "Büyük güncellemelerden önce manuel müdahale gerekip gerekmediğini buradan öğrenirsin.", []string{"haber", "news", "uyarı", "güncelleme öncesi", "arch"}, 0},
	{"bakim", "Sistem bütünlüğünü doğrula", "sudo pacman -Qkk", "Kurulu paketlerin dosyaları bozulmuş mu kontrol eder.", []string{"bütünlük", "bozuk", "doğrula", "kontrol"}, 0},

	// ---------------- METIN ----------------
	{"metin", "Dosya içinde metin ara", "grep -rn 'aranan' /yol", "-r alt klasörler, -n satır numarası, -i büyük/küçük harf duyarsız.", []string{"grep", "ara", "metin", "içinde bul", "kelime"}, 0},
	{"metin", "Komut çıktısını filtrele", "komut | grep kelime", "Pipe ( | ) Linux'un en güçlü aracıdır: bir komutun çıktısını diğerine verir.", []string{"pipe", "grep", "filtre", "çıktı"}, 0},
	{"metin", "Metni topluca değiştir", "sed -i 's/eski/yeni/g' dosya", "-i dosyayı doğrudan değiştirir; önce -i'siz deneyip sonucu gör.", []string{"sed", "değiştir", "replace", "toplu"}, 1},
	{"metin", "Terminalde dosya düzenle (nano)", "nano dosya", "Kaydet: Ctrl+O, Enter. Çık: Ctrl+X. Yeni başlayanlar için en kolayı.", []string{"nano", "düzenle", "editör", "yaz", "kaydet"}, 0},
	{"metin", "Terminalde dosya düzenle (vim)", "vim dosya", "i ile yazma moduna gir, Esc ile çık, :wq kaydedip çıkar, :q! kaydetmeden çıkar.", []string{"vim", "nvim", "editör", "çıkamıyorum"}, 0},
	{"metin", "Root gereken dosyayı düzenle", "sudo nano /etc/dosya", "sudo unutulursa 'read-only' uyarısı alırsın.", []string{"root düzenle", "etc", "yetki", "kaydedemiyorum"}, 1},
	{"metin", "Dosyanın son satırlarını canlı izle", "tail -f dosya", "Log dosyalarını izlemek için birebir.", []string{"tail", "son satır", "izle", "log"}, 0},

	// ---------------- KABUK ----------------
	{"kabuk", "Bir komutun ne yaptığını öğren", "man komut   |   komut --help", "man içinde /kelime ile arama yapar, q ile çıkarsın. Daha kısa örnekler için: tldr komut", []string{"yardım", "man", "help", "nasıl kullanılır", "kılavuz"}, 0},
	{"kabuk", "Komut geçmişinde ara", "Ctrl + R", "Yazmaya başla, eski komutu bulur. Tekrar Ctrl+R ile geriye gider.", []string{"geçmiş", "history", "eski komut", "ctrl r"}, 0},
	{"kabuk", "Bir önceki komutu sudo ile tekrarla", "sudo !!", "'permission denied' aldığında hayat kurtarır.", []string{"sudo", "tekrar", "önceki komut", "denied"}, 0},
	{"kabuk", "Kalıcı kısayol (alias) oluştur", "echo \"alias guncelle='sudo pacman -Syu'\" >> ~/.bashrc", "Fish kullanıyorsan ~/.config/fish/config.fish. Sonra terminali yeniden aç.", []string{"alias", "kısayol", "bashrc", "fish", "takma ad"}, 0},
	{"kabuk", "Komutun nerede kurulu olduğunu bul", "which komut   |   whereis komut", "Komut bulunamıyorsa PATH sorunudur.", []string{"which", "nerede", "yol", "path", "command not found"}, 0},
	{"kabuk", "Terminali kapatınca komut durmasın (oturum yönetimi)", "tmux", "tmux içinde çalıştır, Ctrl+B sonra D ile ayrıl, tmux attach ile geri dön.", []string{"tmux", "screen", "arka plan", "oturum", "kopma"}, 0},
	{"kabuk", "Komutu belirli bir dizinde çalıştır", "(cd /yol && komut)", "Parantez sayesinde bulunduğun dizin değişmez.", []string{"subshell", "dizin", "geçici"}, 0},
	{"kabuk", "Çıktıyı dosyaya yaz", "komut > dosya.txt   |   komut >> dosya.txt", "> üzerine yazar, >> sona ekler. Hataları da almak için: komut &> dosya.txt", []string{"çıktı", "yönlendir", "dosyaya yaz", "kaydet", "redirect"}, 0},
	{"kabuk", "Terminali temizle", "clear   veya   Ctrl + L", "Ekran karıştığında.", []string{"temizle", "clear", "ekran"}, 0},
}

// Yeni başlayanlar için günlük ipuçları
var ipuclari = []string{
	"Arch'ta güncelleme her zaman 'sudo pacman -Syu' şeklinde yapılır. Sadece 'pacman -Sy' ile paket kurmak kısmi güncellemeye ve bozuk sisteme yol açar.",
	"Bir komutu yazmadan önce TAB tuşuna bas: dosya adlarını ve komutları otomatik tamamlar. En çok zaman kazandıran alışkanlık budur.",
	"Riskli bir işlem yapmadan önce snapshot al: sudo snapper -c root create -d 'deneme oncesi'. Btrfs'in en büyük avantajı budur.",
	"Bir hata mesajı aldığında onu olduğu gibi aratmak yerine 'journalctl -p 3 -b' ile gerçek hatayı bul.",
	"'rm' komutunun geri dönüşü yoktur. Silmeden önce aynı yolu 'ls' ile listeleyip doğru yerde olduğunu teyit et.",
	"Bir paketin adını bilmiyorsan 'pacman -Ss kelime' ile ara, sonra 'pacman -Si paket' ile ne olduğunu oku.",
	"AUR paketleri resmi değildir. 'paru -S paket' derken çıkan PKGBUILD'e göz atma alışkanlığı edin.",
	"Terminalde takıldığında Ctrl+C çalışan komutu durdurur, Ctrl+D oturumu kapatır, q ise çoğu görüntüleyiciden (man, less) çıkar.",
	"Sistem yavaşladıysa önce 'htop' aç, sonra 'journalctl -p 3 -b' ile hatalara bak. Tahmin etmek yerine ölç.",
	"Güncelleme sonrası .pacnew dosyaları birikir. Ayda bir 'sudo pacdiff' çalıştırıp yapılandırmaları senkron tut.",
	"Disk dolduysa sırasıyla: sudo paccache -rk2, sudo journalctl --vacuum-size=200M, flatpak uninstall --unused, sonra ncdu ile bak.",
	"Wiki senin en iyi arkadaşın: wiki.archlinux.org üzerindeki bilgiler CachyOS için de neredeyse tamamen geçerlidir.",
	"Bir servisin neden çalışmadığını anlamak için 'systemctl status servis' ve ardından 'journalctl -u servis -e' komutlarını sırayla kullan.",
	"Komutu ezberlemeye çalışma; 'man komut' ve 'komut --help' okumayı alışkanlık haline getir. Ustalık ezber değil, arama becerisidir.",
	"CachyOS'ta indirmeler yavaşsa 'sudo cachyos-rate-mirrors' çalıştır. Çoğu yavaşlık probleminin cevabı budur.",
}

// ---------- Yardımcılar ----------

var trReplacer = strings.NewReplacer(
	"ı", "i", "İ", "i", "I", "i",
	"ş", "s", "Ş", "s",
	"ç", "c", "Ç", "c",
	"ğ", "g", "Ğ", "g",
	"ö", "o", "Ö", "o",
	"ü", "u", "Ü", "u",
)

func norm(s string) string {
	return trReplacer.Replace(strings.ToLower(s))
}

func riskEtiketi(r int) string {
	switch r {
	case 2:
		return red + "[TEHLIKELI]" + reset
	case 1:
		return yellow + "[DIKKAT]" + reset
	}
	return green + "[guvenli]" + reset
}

type sonuc struct {
	c     Cmd
	puan  int
}

func ara(sorgu string) []sonuc {
	q := norm(sorgu)
	kelimeler := strings.Fields(q)
	if len(kelimeler) == 0 {
		return nil
	}
	var out []sonuc
	for _, c := range db {
		puan := 0
		anahtarMetin := norm(strings.Join(c.Anahtar, " "))
		isMetin := norm(c.Is)
		komutMetin := norm(c.Komut)
		notMetin := norm(c.Not)
		katMetin := norm(c.Kat)

		for _, k := range kelimeler {
			if len(k) < 2 {
				continue
			}
			switch {
			case strings.Contains(anahtarMetin, k):
				puan += 10
			case strings.Contains(isMetin, k):
				puan += 7
			case strings.Contains(komutMetin, k):
				puan += 5
			case strings.Contains(katMetin, k):
				puan += 4
			case strings.Contains(notMetin, k):
				puan += 2
			}
		}
		// tam ifade eşleşmesi bonusu
		if strings.Contains(isMetin, q) || strings.Contains(anahtarMetin, q) {
			puan += 8
		}
		if puan > 0 {
			out = append(out, sonuc{c, puan})
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].puan > out[j].puan })
	return out
}

func yazKomut(c Cmd, n int) {
	baslik := fmt.Sprintf("%s%d) %s%s", bold, n, c.Is, reset)
	fmt.Printf("\n%s  %s %s(%s)%s\n", baslik, riskEtiketi(c.Risk), dim, c.Kat, reset)
	fmt.Printf("   %s%s%s\n", cyan+bold, c.Komut, reset)
	if c.Not != "" {
		fmt.Printf("   %s%s%s\n", dim, c.Not, reset)
	}
}

func sonuclariYaz(res []sonuc, limit int) {
	if len(res) == 0 {
		fmt.Printf("\n%sBunu bulamadim.%s Baska kelimelerle dene (ornek: \"wifi\", \"paket sil\", \"disk dolu\").\n", yellow, reset)
		fmt.Printf("%sKategorileri gormek icin: /kat%s\n", dim, reset)
		return
	}
	if len(res) > limit {
		res = res[:limit]
	}
	for i, r := range res {
		yazKomut(r.c, i+1)
	}
	fmt.Println()
}

func kategorileriYaz() {
	fmt.Printf("\n%sKategoriler%s (kullanim: /kat paket)\n\n", bold, reset)
	keys := make([]string, 0, len(kategoriler))
	for k := range kategoriler {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		sayi := 0
		for _, c := range db {
			if c.Kat == k {
				sayi++
			}
		}
		fmt.Printf("  %s%-10s%s %s %s(%d komut)%s\n", green, k, reset, kategoriler[k], dim, sayi, reset)
	}
	fmt.Println()
}

func kategoriYaz(kat string) {
	kat = norm(strings.TrimSpace(kat))
	var bulunan []Cmd
	for _, c := range db {
		if norm(c.Kat) == kat {
			bulunan = append(bulunan, c)
		}
	}
	if len(bulunan) == 0 {
		fmt.Printf("\n%sBoyle bir kategori yok.%s\n", yellow, reset)
		kategorileriYaz()
		return
	}
	fmt.Printf("\n%s=== %s ===%s\n", bold, strings.ToUpper(kat), reset)
	for i, c := range bulunan {
		yazKomut(c, i+1)
	}
	fmt.Println()
}

func quiz(rng *rand.Rand, okuyucu *bufio.Scanner) {
	fmt.Printf("\n%sALISTIRMA MODU%s - 5 soru. Cevabi dusun, Enter'a bas, dogrusunu gor.\n", bold, reset)
	idx := rng.Perm(len(db))
	adet := 5
	if len(idx) < adet {
		adet = len(idx)
	}
	for i := 0; i < adet; i++ {
		c := db[idx[i]]
		fmt.Printf("\n%sSoru %d:%s %s\n", bold+blue, i+1, reset, c.Is)
		fmt.Printf("%s   (cevabi dusun ve Enter'a bas)%s ", dim, reset)
		if !okuyucu.Scan() {
			return
		}
		fmt.Printf("   %s%s%s\n", cyan+bold, c.Komut, reset)
		if c.Not != "" {
			fmt.Printf("   %s%s%s\n", dim, c.Not, reset)
		}
	}
	fmt.Printf("\n%sBitti. Tekrar icin /quiz yaz.%s\n\n", green, reset)
}

func gunlukIpucu(rng *rand.Rand) {
	fmt.Printf("\n%sGUNUN IPUCLARI%s\n\n", bold, reset)
	idx := rng.Perm(len(ipuclari))
	adet := 3
	if len(idx) < adet {
		adet = len(idx)
	}
	for i := 0; i < adet; i++ {
		fmt.Printf("  %s*%s %s\n\n", green, reset, ipuclari[idx[i]])
	}
}

func yardim() {
	fmt.Printf(`
%sNASIL KULLANILIR%s

  Ne yapmak istedigini normal Turkce yaz, komutu bulayim:
    %s> wifi baglanmak istiyorum%s
    %s> disk dolu ne yapayim%s
    %s> paket nasil silinir%s

%sOzel komutlar%s
  /kat            kategorileri listeler
  /kat paket      o kategorideki tum komutlari gosterir
  /quiz           5 soruluk alistirma yapar
  /ipucu          rastgele ogrenme ipuclari verir
  /tehlike        dikkat gerektiren komutlari listeler
  /yardim         bu ekran
  /cikis          programdan cikar (Ctrl+D de olur)

%sTerminalden tek seferlik kullanim%s
  komutcu wifi baglan
  komutcu --kat disk

`, bold, reset, cyan, reset, cyan, reset, cyan, reset, bold, reset, bold, reset)
}

func tehlikeliler() {
	fmt.Printf("\n%s%sDIKKAT GEREKTIREN KOMUTLAR%s - bunlari calistirmadan once iki kez dusun.\n", bold, red, reset)
	n := 0
	for _, c := range db {
		if c.Risk >= 2 {
			n++
			yazKomut(c, n)
		}
	}
	fmt.Printf("\n%sGenel kural: geri donusu olmayan bir sey yapmadan once snapshot al:%s\n", yellow, reset)
	fmt.Printf("   %ssudo snapper -c root create -d 'yedek'%s\n\n", cyan+bold, reset)
}

func banner() {
	fmt.Printf(`
%s +----------------------------------------------+
 |   KOMUTCU - CachyOS / Arch komut asistani    |
 +----------------------------------------------+%s
 %s%d komut, %d kategori. Tamamen offline calisir.%s

 Ne yapmak istedigini Turkce yaz. Ornek: %s"wifi baglan"%s
 Yardim icin %s/yardim%s, cikmak icin %s/cikis%s

`, blue, reset, dim, len(db), len(kategoriler), reset, cyan, reset, green, reset, green, reset)
}

func main() {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Tek seferlik (argümanlı) kullanım
	if len(os.Args) > 1 {
		arg1 := os.Args[1]
		switch arg1 {
		case "--kat", "-k":
			if len(os.Args) > 2 {
				kategoriYaz(os.Args[2])
			} else {
				kategorileriYaz()
			}
		case "--yardim", "-h", "--help":
			yardim()
		case "--ipucu":
			gunlukIpucu(rng)
		default:
			sonuclariYaz(ara(strings.Join(os.Args[1:], " ")), 5)
		}
		return
	}

	banner()
	okuyucu := bufio.NewScanner(os.Stdin)
	okuyucu.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for {
		fmt.Printf("%s>%s ", green+bold, reset)
		if !okuyucu.Scan() {
			fmt.Printf("\n%sGorusuruz. Bol pratik!%s\n", dim, reset)
			return
		}
		satir := strings.TrimSpace(okuyucu.Text())
		if satir == "" {
			continue
		}

		if strings.HasPrefix(satir, "/") {
			parcalar := strings.Fields(satir)
			switch parcalar[0] {
			case "/cikis", "/q", "/exit":
				fmt.Printf("%sGorusuruz. Bol pratik!%s\n", dim, reset)
				return
			case "/yardim", "/h":
				yardim()
			case "/kat":
				if len(parcalar) > 1 {
					kategoriYaz(strings.Join(parcalar[1:], " "))
				} else {
					kategorileriYaz()
				}
			case "/quiz":
				quiz(rng, okuyucu)
			case "/ipucu":
				gunlukIpucu(rng)
			case "/tehlike":
				tehlikeliler()
			default:
				fmt.Printf("%sBilinmeyen komut. /yardim yaz.%s\n", yellow, reset)
			}
			continue
		}

		sonuclariYaz(ara(satir), 5)
	}
}
