# Maintainer: Steven Speek <slspeek@gmail.com>

pkgname=apod-bg
pkgver=1.2
pkgrel=1
pkgdesc="Automatically sets your background to the new Astronomy Picture Of the Day"
arch=('x86_64' 'i686')
url="https://github.com/slspeek/apod-bg"
license=('gpl3')
makedepends=('git' 'go')
depends=('xdg-utils')
optdepends=(
'feh: bare window-manager support'
'pcmanfm: lxde support'
)

options=('!strip' '!emptydirs')
_gourl=github.com/slspeek/apod-bg

build() {
	GOPATH="$srcdir" go get -fix -v -x ${_gourl}
}


package() {
	mkdir -p "$pkgdir/usr/bin"
	install -p -m755 "$srcdir/bin/"* "$pkgdir/usr/bin"

	for f in LICENSE COPYING LICENSE.* COPYING.*; do
		if [ -e "$srcdir/src/$_gourl/$f" ]; then
			install -Dm644 "$srcdir/src/$_gourl/$f" \
				"$pkgdir/usr/share/licenses/$pkgname/$f"
		fi
	done
	cd "$srcdir/src/$_gourl"
	install -Dm644 "README.md" "$pkgdir/usr/share/doc/$pkgname/README.md"
	install -Dm644 "i3wm.config" "$pkgdir/usr/share/doc/$pkgname/i3wm.config"
	install -Dm644	"apod-bg.man" "$pkgdir/usr/share/man/man1/apod-bg.1"
}
md5sums=('SKIP')
