function(e, t, r) {
                var i, a, o, c, p, u, d, f, m, g, y, M, k, w = this.map.getView(), S = [], x = [], L = !1, b = !!this.options.fleet_mode && this.options.fleet_names, _ = this.options.names && t >= 7 || t > A;
                this.options.fleet_mode && b && (_ = !0);
                for (var I = _ ? 1 : 0, P = t > E, F = !1, C = e.byteLength, D = 4; D < C; ) {
                    k = 0;
                    var V = e.getInt16(D)
                      , z = (240 & V) >> 4
                      , G = (16128 & V) >> 8
                      , O = void 0;
                    if (t > 6)
                        switch (49152 & V) {
                        case 49152:
                            O = 2;
                            break;
                        case 32768:
                            O = 0;
                            break;
                        default:
                            O = 1
                        }
                    else
                        O = 1;
                    D += 2,
                    c = e.getInt32(D),
                    D += 4,
                    F = c === r,
                    u = e.getInt32(D) / U,
                    D += 4,
                    p = e.getInt32(D) / U,
                    D += 4,
                    y = new s.geom.Point(s.proj.transform([p, u], "EPSG:4326", v)),
                    F && (
                      a = e.getInt16(D) / 10,
                      D += 2,
                      o = e.getInt16(D) / 10,
                      D += 2
                    );
                    var Z = e.getInt8(D);
                    if ((D += 1) + Z > C)
                        break;
                    M = e.getUTF8String(D, Z),
                    D += Z,
                    "" == M && (M = c.toString()),
                    F && (
                      k = e.getInt32(D),
                      D += 4
                    );
                    var j = 0;
                    if (
                      P && (
                        d = e.getInt16(D),
                        D += 2,
                        f = e.getInt16(D),
                        D += 2,
                        m = e.getInt16(D),
                        D += 2,
                        g = e.getInt16(D),
                        D += 2,
                        j = e.getInt16(D),
                        D += 2,
                        d + f > 0 && m + g > 0 && d + f <= 500 && m < 63 && g < 63
                      )
                    ) {
                          var N = B(u, p, d, f, m, g, j >= 0 && j <= 360 ? j : G < 32 ? Math.floor(11.25 * G) : -1, z);
                          null !== N && x.push(N)
                    }
                    var X = 1 & V
                      , R = 0 != (2 & V)
                      , Y = 0;
                    R && (
                      P ? Y = j : (
                        Y = e.getInt16(D),
                        D += 2
                      )
                    ),
                    i = new T(y,M,c,a,o,!1,k,X,R,F,Y,z,G,O,I),
                    S.push(i),
                    c === this.trackMMSI && this.addCurrentTrackPoint(y.getCoordinates(), a, o, k),
                    F && (this.selectedShipMarker = i,
                    L = !0,
                    (4 & V) > 0 ? (this.oldShip.shown = !0,
                    this.displayOldShip(c, M, k),
                    S.push(this.oldShip.feature)) : this.oldShip.shown = !1)
                }
                if (
                  this.markers.clear(),
                  this.markers.addFeatures(S),
                  this.shapes.clear(),
                  x.length > 0 && this.shapes.addFeatures(x),
                  this.extraMarkers.clear(),
                  L
                ) {
                    this.selectionMarker.setGeometry(this.selectedShipMarker.getGeometry()),
                    this.extraMarkers.addFeature(this.selectedShipMarker),
                    this.markers.addFeature(this.selectionMarker),
                    this.selectionMarkerAdded = !0,
                    this.track_request && (
                      this.getShipTrackW(),
                      this.track_request = !1
                    );
                    var q = w.calculateExtent(this.map.getSize())
                      , H = this.selectedShipMarker.getGeometry().getCoordinates()
                      , W = s.extent.containsCoordinate(q, H);
                    this.options.locate_ship_request ? (
                      n.default.emit(
                        "ter-ship", {
                          mmsi: r,
                          sar: this.selectedShipMarker.get("sar"),
                          feature: this.selectedShipMarker,
                          zoom: t
                        }
                      ),
                      1 === this.lsreq ? (
                        l.default.dispatch(h.mapDidLocateShip()),
                        this.lsreq = 0
                      ) : l.default.dispatch(h.mapDidSelectShip()),
                      this.options.locate_ship_request = !1
                    ) : window.__sipanel.updateDualBtn(W, t)
                } else
                    this.selectionMarkerAdded = !1,
                    this.oldShip.shown = !1
            }
