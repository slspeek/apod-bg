# Maintainer: Steven Speek <slspeek@gmail.com>

pkgname=apod-bg-git
pkgver=65.ad43df6
pkgrel=1
pkgdesc="Automatically sets your background to the new Astronomy Picture Of the Day"
url="https://github.com/slspeek/apod-bg"
arch=('x86_64' 'i686')
license=('gpl3')
makedepends=('git' 'go')
optdepends=(
'feh: bare window-manager support'
'pcmanfm: lxde support'
'dunst: for receiving notifications'
)
options=('!strip' '!emptydirs')
source=('git+https://github.com/slspeek/apod-bg')
md5sums=('SKIP')

_gourl=github.com/slspeek/apod-bg
_gitname="apod-bg"
pkgver() {
  cd $_gitname
  printf "%s.%s" "$(git rev-list --count HEAD)" "$(git rev-parse --short HEAD)"
}

build() {
	GOPATH="$srcdir" go get -fix -t -v -x ${_gourl}
	GOPATH="$srcdir" go get -fix -v -x github.com/stretchr/testify/assert
}

check() {
	GOPATH="$srcdir" go test -v -x ${_gourl}/...
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
