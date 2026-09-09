package roaring64

// to run just these tests: go test -run TestFastAggregations*

import (
	"fmt"
	"math"
	"math/rand"
	"slices"
	"testing"

	"github.com/RoaringBitmap/roaring/v2"
	"github.com/stretchr/testify/assert"
)

func TestFastAggregationsAdvanced_run(t *testing.T) {
	rb1 := NewBitmap()
	rb2 := NewBitmap()
	rb3 := NewBitmap()
	for i := uint64(500); i < 75000; i++ {
		rb1.Add(i)
	}
	for i := uint64(0); i < 1000000; i += 7 {
		rb2.Add(i)
	}
	for i := uint64(0); i < 1000000; i += 1001 {
		rb3.Add(i)
	}
	for i := uint64(1000000); i < 2000000; i += 1001 {
		rb1.Add(i)
	}
	for i := uint64(1000000); i < 2000000; i += 3 {
		rb2.Add(i)
	}
	for i := uint64(1000000); i < 2000000; i += 7 {
		rb3.Add(i)
	}
	rb1.RunOptimize()
	rb1.Or(rb2)
	rb1.Or(rb3)
	bigand := And(And(rb1, rb2), rb3)

	assert.True(t, FastOr(rb1, rb2, rb3).Equals(rb1))
	assert.True(t, FastAnd(rb1, rb2, rb3).Equals(bigand))
}

func TestIssue330_64bits(t *testing.T) {
	var values = [][]uint64{
		{1448147, 1331557, 1166809, 1404655, 1404657, 1448993, 1448994, 1555026, 1568981, 1166795, 1578735, 1456755, 1581128, 1166754, 1357064, 1166799, 1581142, 1166774, 1549034, 1090425, 1061936, 1581118, 1568668, 1456470, 1396063, 1597976, 1458021, 1344102, 1428259, 1166742, 1332378, 1456750, 1313881, 1371860, 1166770, 1513470, 1456761, 1520695, 1322567, 1456765, 1457788, 1166816, 1432713, 1581004, 1451025, 1166729, 1587500, 1581022, 1166707, 1489623, 1581108, 1547596, 1166727, 1345858, 1166741, 1473887, 1581152, 1581114, 1259737, 1434713, 1456740, 1492705, 1316224, 1448997, 1481940, 1456767, 1467183, 1576718, 1458286, 1475626, 1166785, 1428366, 1303084, 1061926, 1388553, 1453950, 1400529, 1581133, 1166713, 1581166, 1279364, 1322319, 1581027, 1166708, 1442325, 1510314, 1166761, 1404658, 1062733, 1166764, 1431819, 1568982, 1322271, 1427065, 1374050, 1166721, 1166714, 1321056, 1303185, 1329366, 1531398, 1071494, 1476413, 1373526, 1166793, 1404659, 1525886, 1166735, 1593361, 1496990, 1166748, 1366912, 1541272, 1166697, 1432913, 1559279, 1456736, 1315451, 1365178, 1068808, 1166768, 1581107, 1345349, 1166792, 1316413, 1449633, 1456758, 1567180, 1448998, 1423954, 1458607, 1480406, 1493217, 1469065, 1581164, 1581024, 1486803, 1550949, 1166803, 1166783, 1072253, 1499822, 1166724, 1559280, 1166732, 1488319, 1166796, 1062666, 1581165, 1507483, 1544041, 1483167, 1315400, 1166762, 1404660, 1581156, 1166786, 1499824, 1340819, 1166775, 1166710, 1473917, 1525946, 1166722, 1457787, 1166752, 1581158, 1500566, 1166736, 1581005, 1525788, 1166700, 1166720, 1341004, 1356523, 1259716, 1166815, 1166734, 1456734, 1393359, 1315351, 1166808, 1589134, 1166749, 1405717, 1386258, 1166733, 1166738, 1316508, 1510168, 1434714, 1341261, 1499823, 1166728, 1089511, 1166790, 1564869, 1316551, 1356413, 1401448, 1166703, 1456762, 1473571, 1405729, 1340780, 1263511, 1464349, 1166810, 1314421, 1519020, 1581006, 1514759, 1468171, 1320091, 1522288, 1456753, 1568393, 1581110, 1530461, 1456744, 1166750, 1518612, 1448325, 1314799, 1166696, 1166813, 1316622, 1489636, 1456741, 1166699, 1529489, 1481939, 1347707, 1448995, 1320330, 1466667, 1166739, 1166787, 1581009, 1323304, 1428380, 1456752, 1450469, 1496415, 1438461, 1450731, 1529496, 1166731, 1581141, 1581103, 1166706, 1064607, 1587650, 1474681, 1064608, 1166756, 1581015, 1573170, 1166801, 1581130, 1549033, 1166694, 1166800, 1593359, 1581155, 1540878, 1599952, 1538583, 1060736, 1166771, 1166711, 1166791, 1475206, 1166755, 1166798, 1455255, 1456751, 1464345, 1370294, 1401160, 1530295, 1573169, 1166804, 1061935, 1090681, 1349316, 1447675, 1449634, 1558271, 1581113, 1587248, 1540108, 1460873, 1562278, 1166788, 1530457, 1493594, 1456754, 1166753, 1320442, 1581153, 1166758, 1166737, 1322566, 1588195, 1166746, 1166777, 1492640, 1322682, 1166698, 1528367, 1445599, 1581026, 1166751, 1445781, 1319119, 1354380, 1581143, 1460447, 1166797, 1500720, 1369579, 1166806, 1581147, 1464393, 1492632, 1166780, 1448996, 1166712, 1593362, 1581151, 1166807, 1478408, 1322672, 1581137, 1581002, 1456764, 1405048, 1166766, 1063491, 1166811, 1296758, 1384972, 1314177, 1166730, 1066109, 1514764, 1567450, 1581119, 1573634, 1415662, 1563521, 1059518, 1456738, 1581157, 1550678, 1166726, 1166719, 1581013, 1495998, 1371312, 1584782, 1456471, 1166778, 1458759, 1060733, 1166784, 1581028, 1166781, 1583260, 1522601, 1491470, 1166717, 1319114, 1541228, 1456756, 1448992, 1166744, 1549031, 1570164, 1489634, 1370005, 1475624, 1514141, 1475625, 1525883, 1166817, 1166759, 1488332, 1313675, 1259957, 1581011, 1166725, 1483313, 1417102, 1166740, 1260680, 1166779, 1581134, 1315627, 1430811, 1525947, 1166763, 1435903, 1389401, 1351767, 1343593, 1537920, 1480377, 1520699, 1319223, 1581148, 1589060, 1060735, 1468871, 1087817, 1259068, 1581146, 1456759, 1404656, 1166776, 1166695, 1324844, 1589473, 1374167, 1166769, 1166789, 1456871, 1477673, 1166805, 1581106, 1379727, 1166723, 1166767, 1166745, 1313278, 1581124, 1451652, 1405710, 1408256, 1360274, 1325001, 1581136, 1166718, 1072246, 1410198, 1456746, 1557149, 1166760, 1308054, 1581159, 1581131, 1467536, 1166757, 1434322, 1541831, 1418436, 1574603, 1437326, 1166814, 1584856, 1388750, 1166743, 1372695, 1593360, 1166765, 1166794, 1474005, 1166782, 1456748, 1539639, 1056239, 1166802, 1166747, 1210989},
		{1332290, 1447737, 1549291, 1187244, 1598129, 1579851, 1424171, 1538999, 1445358, 1586739, 1575050, 1437657, 1366343, 1062799, 1421550, 1460317, 1474875, 1060737, 1330773, 1447797, 1348633, 1559437, 1556214, 1187305, 1187234, 1187240, 1464660, 1567794, 1187260, 1260646, 1311938, 1573195, 1318525, 1484524, 1456152, 1087954, 1556007, 1187265, 1460920, 1485316, 1447849, 1447744, 1474001, 1537891, 1478211, 1313292, 1488405, 1187239, 1378814, 1343620, 1500498, 1567809, 1435838, 1506575, 1368282, 1447441, 1598101, 1067076, 1572997, 1598102, 1332697, 1324653, 1561437, 1187290, 1059945, 1187278, 1457187, 1430003, 1450643, 1447436, 1260650, 1473393, 1187247, 1087323, 1324967, 1187291, 1480771, 1472729, 1555881, 1187253, 1456481, 1452672, 1447435, 1378603, 1574771, 1187235, 1417857, 1568706, 1576739, 1534662, 1410189, 1587745, 1473791, 1187308, 1447730, 1328158, 1409164, 1506591, 1500147, 1433961, 1483709, 1187227, 1456479, 1562595, 1314333, 1187281, 1598012, 1415200, 1447791, 1371379, 1598019, 1435836, 1457188, 1457351, 1187248, 1417111, 1187289, 1187252, 1187257, 1316665, 1473464, 1187263, 1447732, 1520371, 1525651, 1598177, 1406947, 1465787, 1524659, 1324213, 1418439, 1575816, 1522696, 1187286, 1497821, 1333179, 1187282, 1335988, 1548952, 1556066, 1314993, 1187276, 1420503, 1187301, 1456468, 1423939, 1598089, 1504357, 1343247, 1437659, 1525768, 1539279, 1307385, 1187275, 1524305, 1332938, 1516498, 1303247, 1304237, 1187238, 1385283, 1595495, 1187300, 1187241, 1061740, 1316383, 1187307, 1062037, 1538693, 1454292, 1447731, 1187272, 1561442, 1187268, 1567150, 1597966, 1447745, 1598178, 1187262, 1464067, 1325394, 1537893, 1332693, 1479200, 1522335, 1589378, 1450573, 1399161, 1421274, 1561501, 1187232, 1187302, 1258469, 1331600, 1447740, 1187242, 1328147, 1264069, 1187294, 1299943, 1598013, 1526975, 1260604, 1487518, 1187229, 1487617, 1354087, 1456595, 1462047, 1561438, 1598363, 1332691, 1424655, 1567105, 1574774, 1598035, 1526981, 1384038, 1475987, 1343587, 1447437, 1454912, 1382215, 1447739, 1456512, 1447779, 1187283, 1440988, 1187293, 1187298, 1574754, 1354772, 1598018, 1447429, 1598181, 1447738, 1598273, 1312197, 1574752, 1572995, 1526127, 1473908, 1437660, 1447743, 1362262, 1456513, 1539280, 1348625, 1415878, 1332694, 1471020, 1432462, 1058088, 1526710, 1371788, 1187288, 1537984, 1316874, 1187270, 1333565, 1187292, 1447796, 1187311, 1187237, 1187231, 1574755, 1553822, 1522019, 1447418, 1187269, 1332692, 1447735, 1529638, 1468154, 1328031, 1447733, 1447402, 1593884, 1332696, 1560622, 1564819, 1538967, 1315756, 1328338, 1598113, 1324212, 1449895, 1567793, 1260629, 1430010, 1187266, 1187256, 1312754, 1449417, 1595494, 1529054, 1187261, 1187306, 1526976, 1425490, 1366922, 1527390, 1187299, 1561510, 1319222, 1187250, 1057262, 1457999, 1332937, 1187243, 1556213, 1278602, 1546839, 1187296, 1548950, 1580141, 1187303, 1187255, 1525650, 1572998, 1576740, 1187267, 1464664, 1440427, 1456467, 1187271, 1187258, 1585428, 1548760, 1342254, 1447793, 1406348, 1500177, 1260644, 1416954, 1323722, 1412713, 1187280, 1187310, 1538015, 1537285, 1187285, 1456482, 1260611, 1490508, 1187274, 1585641, 1416648, 1484655, 1421520, 1347485, 1525652, 1568987, 1526974, 1314375, 1187246, 1455623, 1488117, 1445025, 1447401, 1478237, 1561440, 1187287, 1561434, 1509337, 1451859, 1599630, 1348639, 1449436, 1361844, 1464661, 1263064, 1526973, 1187279, 1562080, 1354770, 1454521, 1520719, 1478236, 1526972, 1423948, 1334866, 1325026, 1438275, 1422582, 1437646, 1315530, 1458323, 1447795, 1528218, 1187254, 1598056, 1417853, 1423514, 1187297, 1187245, 1187264, 1524662, 1187251, 1524660, 1328113, 1187304, 1374767, 1474057, 1187284, 1331601, 1598180, 1062814, 1488818, 1187309, 1087494, 1063499, 1458325, 1187295, 1432336, 1260001, 1597982, 1537147, 1445355, 1595491, 1396111, 1546848, 1474048, 1495251, 1447734, 1464071, 1526978, 1187236, 1526977, 1566267, 1187277, 1421549, 1430015, 1316024, 1332695, 1561433, 1435837, 1087250, 1574753, 1476183, 1325395, 1561432, 1447736, 1500181, 1424164, 1456483, 1187228, 1573384, 1273769, 1598085, 1437661, 1306415, 1407257, 1187249, 1338215, 1458047, 1520791, 1447741, 1537263, 1472490, 1524661, 1061729, 1187273, 1417861, 1470196, 1485881, 1260595, 1538846, 1568762, 1315170, 1500469, 1372455, 1558140, 1425202, 1432702, 1472734, 1187230, 1187312, 1598191, 1569680, 1187233, 1263091, 1447417, 1504429, 1430016, 1435839, 1458324, 1546845, 1575027, 1187259, 1464610},
		{1578810, 1166701, 1335901, 1063397, 1578812, 1526471, 1166702, 1067008, 1412862, 1059750, 1060729, 1166812, 1493768, 1335772, 1336194, 1166772, 1465775, 1166704, 1355314, 1314138, 1060727, 1388260, 1465786, 1565378, 1522096, 1312490, 1319284, 1418573, 1319169, 1060730, 1307995, 1465780, 1060728, 1404061, 1407842, 1326256, 1578754, 1410577, 1060732, 1461338, 1264209, 1166705, 1581273, 1414196, 1565398, 1355106, 1060731, 1412641, 1306618, 1060726, 1356302, 1310896, 1457885, 1497721, 1166709, 1067009, 1310812, 1578800, 1422203, 1484409, 1485278, 1322057, 1369956, 1311089, 1576614, 1355711, 1355798, 1564832, 1304166, 1166773, 1319071, 1578805, 1575869, 1403066},
	}

	bitmaps := []*Bitmap{}

	for _, v := range values {
		bitmap := BitmapOf(v...)
		bitmap.RunOptimize()
		bitmaps = append(bitmaps, bitmap)
		slices.Sort(v)
		arrayResult := bitmap.ToArray()
		assert.Equal(t, arrayResult, v)
	}
	assert.Equal(t, FastAnd(bitmaps[0], bitmaps[1]).GetCardinality(), uint64(0))
	assert.Equal(t, FastAnd(bitmaps[0], bitmaps[2]).GetCardinality(), uint64(0))
	assert.Equal(t, FastAnd(bitmaps[1], bitmaps[2]).GetCardinality(), uint64(0))
	assert.True(t, FastOr(bitmaps[1], bitmaps[2]).Equals(FastOr(bitmaps[2], bitmaps[1])))
	assert.True(t, FastOr(bitmaps[1], bitmaps[0]).Equals(FastOr(bitmaps[0], bitmaps[1])))
	assert.Equal(t, FastOr(bitmaps[0], bitmaps[1], bitmaps[2]).GetCardinality(), uint64(1040))
	assert.Equal(t, FastOr(bitmaps[2], bitmaps[1], bitmaps[0]).GetCardinality(), uint64(1040))
	assert.True(t, FastOr(bitmaps[2], bitmaps[0]).Equals(FastOr(bitmaps[0], bitmaps[2])))
	assert.True(t, FastOr(bitmaps[2], bitmaps[1], bitmaps[0]).Equals(FastOr(bitmaps[1], bitmaps[2], bitmaps[0])))
	assert.True(t, FastOr(bitmaps[2], bitmaps[1], bitmaps[0]).Equals(FastOr(bitmaps[0], bitmaps[1], bitmaps[2])))
	assert.Equal(t, FastOr(bitmaps[2], bitmaps[1], bitmaps[0]).GetCardinality(), uint64(1040))
	assert.Equal(t, FastOr(bitmaps[0], bitmaps[1], bitmaps[2]).GetCardinality(), uint64(1040))
}

func TestFastRangeAggregations(t *testing.T) {
	andFunc := func(bitmaps ...*Bitmap) *Bitmap {
		return FastAndRange(0, math.MaxUint64, bitmaps...)
	}
	orFunc := func(bitmaps ...*Bitmap) *Bitmap {
		return FastOrRange(0, math.MaxUint64, bitmaps...)
	}
	testAggregations(t, andFunc, orFunc, nil)
}

// clipToRange removes every value of b outside [rangeStart, rangeEnd).
// RemoveRange cannot express an end past math.MaxUint64, so that value is
// removed on its own; no half-open range selects it.
func clipToRange(b *Bitmap, rangeStart, rangeEnd uint64) {
	if rangeStart >= rangeEnd {
		b.Clear()
		return
	}
	b.RemoveRange(0, rangeStart)
	b.RemoveRange(rangeEnd, math.MaxUint64)
	b.Remove(math.MaxUint64)
}

// referenceOrRange is the materialize-then-clip computation FastOrRange
// replaces.
func referenceOrRange(rangeStart, rangeEnd uint64, bitmaps ...*Bitmap) *Bitmap {
	answer := FastOr(bitmaps...)
	clipToRange(answer, rangeStart, rangeEnd)
	return answer
}

// referenceAndRange is the materialize-then-clip computation FastAndRange
// replaces.
func referenceAndRange(rangeStart, rangeEnd uint64, bitmaps ...*Bitmap) *Bitmap {
	answer := FastAnd(bitmaps...)
	clipToRange(answer, rangeStart, rangeEnd)
	return answer
}

// roaringArray64Snapshot records everything a range aggregation must leave
// alone in an input: the copy-on-write setting, the keys, the container
// pointers, their contents and copy-on-write settings, and the per-container
// copy-on-write flags. The flags inside the 32-bit containers are not
// reachable from this package; the 32-bit tests cover them.
type roaringArray64Snapshot struct {
	copyOnWrite   bool
	keys          []uint32
	containers    []*roaring.Bitmap
	contents      []*roaring.Bitmap
	containersCOW []bool
	flags         []bool
}

func snapshotRoaringArray64(ra *roaringArray64) roaringArray64Snapshot {
	s := roaringArray64Snapshot{copyOnWrite: ra.copyOnWrite}
	s.keys = append(s.keys, ra.keys...)
	s.containers = append(s.containers, ra.containers...)
	s.flags = append(s.flags, ra.needCopyOnWrite...)
	for _, c := range ra.containers {
		// Clone would flag the containers of a copy-on-write source, so
		// copy through the serialized form instead.
		buf, err := c.ToBytes()
		if err != nil {
			panic(err)
		}
		content := roaring.New()
		if err := content.UnmarshalBinary(buf); err != nil {
			panic(err)
		}
		s.contents = append(s.contents, content)
		s.containersCOW = append(s.containersCOW, c.GetCopyOnWrite())
	}
	return s
}

func assertRoaringArray64Unchanged(t *testing.T, s roaringArray64Snapshot, ra *roaringArray64) {
	t.Helper()
	assert.Equal(t, s.copyOnWrite, ra.copyOnWrite, "copy-on-write setting changed")
	if !assert.Equal(t, len(s.keys), len(ra.keys), "number of containers changed") {
		return
	}
	for i := range s.keys {
		assert.Equal(t, s.keys[i], ra.keys[i], "key %d changed", i)
		assert.Equal(t, s.flags[i], ra.needCopyOnWrite[i], "copy-on-write flag %d changed", i)
		assert.True(t, s.containers[i] == ra.containers[i], "container %d was replaced", i)
		assert.True(t, s.contents[i].Equals(ra.containers[i]), "container %d was modified", i)
		assert.Equal(t, s.containersCOW[i], ra.containers[i].GetCopyOnWrite(), "container %d copy-on-write setting changed", i)
	}
}

// checkRangeAggregation runs both range aggregations of bitmaps over
// [rangeStart, rangeEnd), checks that the inputs are left untouched, and
// compares the validated results with the reference. The reference runs last
// because FastOr and FastAnd clone a single input, and Clone flags the
// containers of a copy-on-write source.
func checkRangeAggregation(t *testing.T, rangeStart, rangeEnd uint64, bitmaps ...*Bitmap) {
	t.Helper()
	snapshots := make([]roaringArray64Snapshot, len(bitmaps))
	for i, b := range bitmaps {
		snapshots[i] = snapshotRoaringArray64(&b.highlowcontainer)
	}
	gotOr := FastOrRange(rangeStart, rangeEnd, bitmaps...)
	gotAnd := FastAndRange(rangeStart, rangeEnd, bitmaps...)
	for i, b := range bitmaps {
		assertRoaringArray64Unchanged(t, snapshots[i], &b.highlowcontainer)
	}

	assert.NoError(t, gotOr.Validate())
	assert.False(t, gotOr.GetCopyOnWrite())
	wantOr := referenceOrRange(rangeStart, rangeEnd, bitmaps...)
	assert.True(t, gotOr.Equals(wantOr), "FastOrRange(%d, %d): got %s, want %s", rangeStart, rangeEnd, gotOr, wantOr)

	assert.NoError(t, gotAnd.Validate())
	assert.False(t, gotAnd.GetCopyOnWrite())
	wantAnd := referenceAndRange(rangeStart, rangeEnd, bitmaps...)
	assert.True(t, gotAnd.Equals(wantAnd), "FastAndRange(%d, %d): got %s, want %s", rangeStart, rangeEnd, gotAnd, wantAnd)
}

// rangeFixtures returns bitmaps spread over several 32-bit containers: array,
// bitmap and run containers inside container 0, a run across the boundary of
// containers 1 and 2, a mixed bitmap with a container gap, values at the top
// of the value space including math.MaxUint64, and a far-away container.
func rangeFixtures(cow bool) []*Bitmap {
	arrays := New()
	for i := uint64(0); i < 3855; i++ {
		arrays.Add(i * 17)
		arrays.Add(1<<16 + i*17)
	}

	bitmaps := New()
	for i := uint64(0); i < 1<<16; i += 3 {
		bitmaps.Add(2<<16 + i)
		bitmaps.Add(1<<32 + i)
	}

	runs := New()
	runs.AddRange(1<<32-70000, 2<<32+5000)
	runs.AddRange(2<<32+1<<20, 2<<32+1<<20+100)
	runs.RunOptimize()

	mixed := New()
	mixed.AddMany([]uint64{5, 10, 15, 1<<32 + 7, 5<<32 + 1})
	for i := uint64(0); i < 1<<16; i += 2 {
		mixed.Add(5<<32 + 3<<16 + i)
	}
	mixed.AddRange(5<<32+4<<16, 5<<32+6<<16)
	mixed.RunOptimize()

	top := New()
	top.AddRange(math.MaxUint64-1000, math.MaxUint64)
	top.Add(math.MaxUint64)
	top.Add(math.MaxUint64 - 70000)
	top.RunOptimize()

	far := New()
	far.AddMany([]uint64{100 << 32, 100<<32 + 1, 200<<32 + 7})

	fixtures := []*Bitmap{arrays, bitmaps, runs, mixed, top, far}
	for _, b := range fixtures {
		b.SetCopyOnWrite(cow)
	}
	if cow {
		fixtures = append(fixtures, fixtures[2].Clone())
	}
	return fixtures
}

func rangeWindows() []struct {
	name       string
	start, end uint64
} {
	return []struct {
		name       string
		start, end uint64
	}{
		{"empty start==end", 5, 5},
		{"empty start>end", 10, 3},
		{"container gap", 6 << 32, 100 << 32},
		{"above every key", 201 << 32, math.MaxUint64},
		{"inside array container", 100, 5000},
		{"inside bitmap container", 2<<16 + 100, 2<<16 + 5000},
		{"inside run", 1<<32 + 100, 1<<32 + 5000},
		{"run straddles both edges", 1<<32 - 50000, 2<<32 + 3000},
		{"across the container boundary", 1<<32 - 100, 1<<32 + 100},
		{"container aligned", 1 << 32, 2 << 32},
		{"single present value", 2<<16 + 3, 2<<16 + 4},
		{"single absent value", 2<<16 + 4, 2<<16 + 5},
		{"top of value space", math.MaxUint64 - 5, math.MaxUint64},
		{"top container", math.MaxUint64 - 100000, math.MaxUint64},
		{"whole", 0, math.MaxUint64},
		{"wide", 1<<16 + 5, 200<<32 + 7},
	}
}

func TestFastRangeAggregationsWindows(t *testing.T) {
	for _, cow := range []bool{false, true} {
		fixtures := rangeFixtures(cow)
		empty := New()
		subsets := map[string][]*Bitmap{
			"none":       {},
			"empty":      {empty},
			"empty+runs": {empty, fixtures[2]},
			"twice":      {fixtures[3], fixtures[3]},
			"all":        fixtures,
		}
		for i, a := range fixtures {
			subsets[fmt.Sprintf("single%d", i)] = []*Bitmap{a}
			for j, b := range fixtures[i+1:] {
				subsets[fmt.Sprintf("pair%d%d", i, i+1+j)] = []*Bitmap{a, b}
			}
		}
		for name, bitmaps := range subsets {
			for _, w := range rangeWindows() {
				t.Run(fmt.Sprintf("cow=%v/%s/%s", cow, name, w.name), func(t *testing.T) {
					checkRangeAggregation(t, w.start, w.end, bitmaps...)
				})
			}
		}
	}
}

func TestFastRangeAggregationsResultIsolated(t *testing.T) {
	// Writes to the results must not reach the inputs and writes to the
	// inputs must not reach the results, even through shared containers.
	for _, cow := range []bool{false, true} {
		for i, b := range rangeFixtures(cow) {
			t.Run(fmt.Sprintf("cow=%v/input%d", cow, i), func(t *testing.T) {
				snapshot := snapshotRoaringArray64(&b.highlowcontainer)
				or := FastOrRange(0, math.MaxUint64, b)
				and := FastAndRange(1<<16, 3<<32, b, b)
				for _, r := range []*Bitmap{or, and} {
					r.Add(1<<32 + 7)
					r.RemoveRange(0, 1<<32+10)
					r.Flip(1<<32, 1<<32+64)
				}
				assertRoaringArray64Unchanged(t, snapshot, &b.highlowcontainer)

				or = FastOrRange(0, math.MaxUint64, b)
				and = FastAndRange(1<<16, 3<<32, b, b)
				orWant := FastOrRange(0, math.MaxUint64, or)
				andWant := FastOrRange(0, math.MaxUint64, and)
				b.Add(1<<32 + 3)
				b.Remove(1 << 32)
				b.AddRange(2<<32+1, 2<<32+3)
				b.RemoveRange(5<<32, 6<<32)
				b.Flip(2<<16, 2<<16+64)
				assert.True(t, or.Equals(orWant))
				assert.True(t, and.Equals(andWant))
			})
		}
	}
}

// randomRangeBitmap builds a bitmap over the given 32-bit containers where
// every 16-bit container is absent, an array, a bitmap or a set of runs at
// random, and applies RunOptimize and copy-on-write at random.
func randomRangeBitmap(rng *rand.Rand, highs []uint64, keys int) *Bitmap {
	b := New()
	var values []uint64
	for _, high := range highs {
		for key := 0; key < keys; key++ {
			base := high<<32 + uint64(key)<<16
			switch rng.Intn(4) {
			case 0:
			case 1:
				values = values[:0]
				for i, n := 0, 1+rng.Intn(3000); i < n; i++ {
					values = append(values, base+uint64(rng.Intn(1<<16)))
				}
				b.AddMany(values)
			case 2:
				values = values[:0]
				for i, n := 0, 5000+rng.Intn(20000); i < n; i++ {
					values = append(values, base+uint64(rng.Intn(1<<16)))
				}
				b.AddMany(values)
			case 3:
				for r, runs := 0, 1+rng.Intn(4); r < runs; r++ {
					start := uint64(rng.Intn(1 << 16))
					end := start + 1 + uint64(rng.Intn(20000))
					if end > 1<<16 {
						end = 1 << 16
					}
					b.AddRange(base+start, base+end)
				}
			}
		}
	}
	if rng.Intn(2) == 0 {
		b.RunOptimize()
	}
	if rng.Intn(2) == 0 {
		b.SetCopyOnWrite(true)
	}
	return b
}

// randomRangeBound returns a range bound that is often on or next to a
// 32-bit or 16-bit container boundary, sometimes at the top of the value
// space, and otherwise uniform over the containers in use.
func randomRangeBound(rng *rand.Rand, highs []uint64, keys int) uint64 {
	high := highs[rng.Intn(len(highs))]
	switch rng.Intn(6) {
	case 0:
		return high<<32 + uint64(rng.Intn(keys+1))<<16
	case 1:
		return high<<32 + uint64(rng.Intn(keys+1))<<16 + uint64(rng.Intn(3))
	case 2:
		return high<<32 + uint64(1+rng.Intn(keys+1))<<16 - uint64(1+rng.Intn(2))
	case 3:
		return high<<32 + uint64(rng.Intn(3)) - uint64(rng.Intn(2))
	case 4:
		return []uint64{math.MaxUint64, math.MaxUint64 - 1, 1 << 40}[rng.Intn(3)]
	}
	return high<<32 + uint64(rng.Intn((keys+1)<<16))
}

func TestFastRangeAggregationsRandom(t *testing.T) {
	rng := rand.New(rand.NewSource(20260909))
	highs := []uint64{0, 1, 2, 7, math.MaxUint32}
	const keys = 3
	for iteration := 0; iteration < 150; iteration++ {
		bitmaps := make([]*Bitmap, 1+rng.Intn(4))
		for i := range bitmaps {
			if i > 0 && rng.Intn(6) == 0 {
				bitmaps[i] = bitmaps[rng.Intn(i)]
				continue
			}
			bitmaps[i] = randomRangeBitmap(rng, highs[:1+rng.Intn(len(highs))], keys)
		}
		start, end := randomRangeBound(rng, highs, keys), randomRangeBound(rng, highs, keys)
		if start > end && rng.Intn(4) != 0 {
			start, end = end, start
		}
		checkRangeAggregation(t, start, end, bitmaps...)
		if t.Failed() {
			t.Fatalf("iteration %d, range [%d, %d), %d inputs", iteration, start, end, len(bitmaps))
		}
	}
}
