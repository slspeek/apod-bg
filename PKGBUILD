# Maintainer: Steven Speek <slspeek@gmail.com>

pkgname=apod-bg
pkgver=1.0
pkgrel=1
pkgdesc="Automatically sets your background to the new Astronomy Picture Of the Day"
arch=('any')
url="https://github.com/slspeek/apod-bg"
license=('gpl3')
makedepends=('git' 'go')
source=("$pkgname"::'git://github.com/slspeek/apod-bg.git')

build() {
	export GOPATH=$(pwd)
	go get github.com/slspeek/apod-bg
}

package() {
	cd bin
	install -Dm755 "$pkgname" "$pkgdir/usr/bin/$pkgname"
	cd "$srcdir/$pkgname"
	install -Dm644 "README.md" "$pkgdir/usr/share/doc/$pkgname/README.md"
}
md5sums=('SKIP')
