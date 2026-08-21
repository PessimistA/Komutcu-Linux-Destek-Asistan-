# Komutcu Linux Destek Asistanı

**CachyOS / Arch Linux için Türkçe terminal komut asistanı.**

Türkçe yazılan bir istek karşılığında doğru komutu, ne işe yaradığını ve nelere
dikkat edilmesi gerektiğini gösterir.

Tamamen **çevrimdışı** çalışır. Yapay zekâ, internet bağlantısı, API anahtarı ya
da harici bağımlılık yoktur; komut veritabanı programın kendi içine gömülüdür.

```
> disk dolu ne yapayım

1) Boş disk alanını gör  [guvenli] (disk)
   df -h
   -h insan okuyabilir birimlerde (GB) gösterir.

2) En çok yer kaplayanları görsel bul  [guvenli] (disk)
   ncdu /
   Önce: sudo pacman -S ncdu. Ok tuşlarıyla gezinilir, disk dolduğunda hayat kurtarır.

3) Snapshot'ları boot menüsüne yaz (Limine)  [guvenli] (disk)
   sudo limine-snapper-sync
   CachyOS'un Limine bootloader'ında snapshot'tan açılış için gerekir.
```

---

## İçindekiler

- [Neden var?](#neden-var)
- [Kurulum](#kurulum)
- [Kullanım](#kullanım)
- [Özel komutlar](#özel-komutlar)
- [Risk etiketleri](#risk-etiketleri)
- [İçerik](#i̇çerik)
- [Arama](#arama)
- [Genişletme](#genişletme)
- [Proje yapısı](#proje-yapısı)
- [Lisans](#lisans)

---

## Neden var?

Arch tabanlı bir sisteme yeni geçen kullanıcı için asıl engel komutları
hatırlamak değil, hangi komutun ne zaman güvenli olduğunu bilmemektir.
`pacman -Sy` ile `pacman -Syu` arasındaki fark sistemi kırabilir; `rm -rf` bir
saniyede geri dönülemez bir hataya dönüşebilir.

komutcu bu boşluğu doldurur: her komutun yanında bir **risk etiketi** ve kısa bir
uyarı notu bulunur. Yalnızca komut verilmez, nedeni de anlatılır.

Go tercih edilmiştir: tek dosyalık ve bağımlılıksız bir ikili üretir, çalışma
zamanı ya da sanal ortam gerektirmez, açılışı anlıktır ve veritabanı programın
içine gömülü olduğu için internet gerekmez.

---

## Kurulum

Üç dosya (`main.go`, `go.mod`, `kur.sh`) aynı klasörde bulunmalıdır:

```bash
cd komutcu
chmod +x kur.sh
./kur.sh
```

Kurulum betiği Go'nun kurulu olup olmadığını denetler, gerekirse `pacman` ile
kurar, projeyi derler ve `/usr/local/bin/komutcu` altına yerleştirir.

Elle kurulum:

```bash
sudo pacman -S go
go build -o komutcu .
sudo install -Dm755 komutcu /usr/local/bin/komutcu
```

**Gereksinim:** Go 1.21+ — yalnızca derleme için; çalıştırmak hiçbir şey
gerektirmez.

---

## Kullanım

### Etkileşimli mod

Asıl kullanım şekli budur:

```bash
komutcu
```

Ardından doğal Türkçe cümleler yazılır:

```
> wifi bağlanmak istiyorum
> disk dolu ne yapayım
> paket nasıl silinir
> ekran kartı sürücüsü
> donmuş programı kapat
```

Çıkış için `/cikis` veya `Ctrl+D`.

### Tek seferlik kullanım

```bash
komutcu wifi baglan
komutcu disk dolu
komutcu --kat cachyos
komutcu --ipucu
komutcu --yardim
```

Türkçe karakter kullanmak şart değildir; arama, Türkçe harfleri karşılıklarına
çevirerek eşleştirir (`ğüşıöç` ↔ `gusioc`).

---

## Özel komutlar

| Komut | İşlevi |
|---|---|
| `/kat` | Kategorileri listeler |
| `/kat paket` | O kategorideki tüm komutları gösterir |
| `/quiz` | 5 soruluk alıştırma yapar |
| `/ipucu` | Rastgele öğrenme ipucu verir |
| `/tehlike` | Geri dönüşü olmayan komutları tek listede gösterir |
| `/yardim` | Yardım ekranı |
| `/cikis` | Çıkış (`Ctrl+D` de olur) |

---

## Risk etiketleri

Her komutun yanında bir etiket bulunur:

| Etiket | Anlamı |
|---|---|
| `[guvenli]` | Doğrudan çalıştırılabilir |
| `[DIKKAT]` | Ne yaptığı anlaşıldıktan sonra çalıştırılmalıdır |
| `[TEHLIKELI]` | Geri dönüşü yoktur; öncesinde snapshot alınmalıdır |

`/tehlike` komutu veritabanındaki bütün `[TEHLIKELI]` kayıtları tek ekranda
toplar; yeni başlayanların ilk okuması gereken liste budur.

---

## İçerik

**14 kategoride 120+ komut:**

| Kategori | Kapsam |
|---|---|
| `paket` | pacman, paru/AUR, flatpak, keyring ve kilit sorunları |
| `sistem` | systemd servisleri, çekirdek, güncelleme |
| `dosya` | Dosya ve dizin işlemleri |
| `disk` | Disk, bölüm, btrfs ve snapshot |
| `ag` | Ağ, wifi, ssh, indirme |
| `surec` | Süreç yönetimi |
| `kullanici` | Kullanıcı, izin, sudo |
| `log` | Log okuma ve hata ayıklama |
| `donanim` | Donanım, sürücü, GPU |
| `cachyos` | CachyOS'a özel araçlar (`chwd`, `cachyos-rate-mirrors`, `cachyos-kernel-manager`) |
| `oyun` | Oyun, Steam, Proton |
| `bakim` | Temizlik ve bakım |
| `metin` | Metin işleme ve editörler |
| `kabuk` | Kabuk hileleri |

---

## Arama

Girilen cümle kelimelere ayrılır ve her komut kaydına puan verilir. Eşleşmenin
nerede olduğuna göre ağırlık değişir:

| Eşleşme yeri | Puan |
|---|---|
| Arama anahtar kelimeleri | 10 |
| Komutun ne işe yaradığı | 7 |
| Komutun kendisi | 5 |
| Kategori adı | 4 |
| Açıklama notu | 2 |
| Tam ifade eşleşmesi (bonus) | +8 |

Sonuçlar puana göre sıralanır. Bu sayede "paket nasıl silinir" gibi doğal bir
cümle doğrudan `pacman -Rns` kaydını üste taşır.

---

## Genişletme

`main.go` içindeki `db` listesine bir satır eklenip yeniden derlemek yeterlidir:

```go
{"paket", "Ne işe yarar", "komut -x", "Açıklama/ipucu", []string{"anahtar", "kelime"}, 0},
```

Sırasıyla: kategori, açıklama, komut, not, arama anahtarları ve risk (`0`
güvenli / `1` dikkat / `2` tehlikeli). Ardından:

```bash
go build -o komutcu . && sudo install -Dm755 komutcu /usr/local/bin/komutcu
```

Yeni bir kategori için `kategoriler` haritasına da bir satır eklenir.

---

## Proje yapısı

```
komutcu/
├── main.go      # tüm program: veri modeli, komut veritabanı, arama, arayüz
├── go.mod       # modül tanımı (bağımlılık yok)
└── kur.sh       # derle + /usr/local/bin altına kur
```

Tek dosya, ~520 satır, standart kütüphane dışında hiçbir paket kullanılmaz.

---

## Lisans

Kişisel kullanım ve öğrenme amaçlı geliştirilmiştir. Komutlar sistemde
değişiklik yapar; özellikle `[DIKKAT]` ve `[TEHLIKELI]` etiketli komutlar
çalıştırılmadan önce ne yaptıkları anlaşılmalıdır.
