package main

import (
	"unsafe"

	. "github.com/linux-surface/iptsd/processing"
	. "github.com/linux-surface/iptsd/protocol"
)

var (
	heatmapBufferCache map[uint16][]byte = make(map[uint16][]byte)
	points             []ContactPoint    = make([]ContactPoint, 10)
)

func IptsTouchHandleHeatmap(ipts *IptsContext, heatmap Heatmap) error {
	touch := ipts.Devices.Touch
	count := heatmap.ContactPoints(points)

	for i := 0; i < count; i++ {
		p := heatmap.GetCoords(points[i])

		touch.Emit(EV_ABS, ABS_MT_SLOT, int32(i))
		touch.Emit(EV_ABS, ABS_MT_TRACKING_ID, int32(i))
		touch.Emit(EV_ABS, ABS_MT_POSITION_X, int32(p.X))
		touch.Emit(EV_ABS, ABS_MT_POSITION_Y, int32(p.Y))
	}

	for i := count; i < len(points); i++ {
		touch.Emit(EV_ABS, ABS_MT_SLOT, int32(i))
		touch.Emit(EV_ABS, ABS_MT_TRACKING_ID, -1)
	}

	err := touch.Emit(EV_SYN, SYN_REPORT, 0)
	if err != nil {
		return err
	}

	return nil
}

func IptsTouchHandleInput(ipts *IptsContext, frame IptsPayloadFrame) error {
	size := uint32(0)
	hm := Heatmap{}

	for size < frame.Size {
		report, err := ipts.Protocol.ReadReport()
		if err != nil {
			return err
		}

		size += uint32(report.Size) + uint32(unsafe.Sizeof(report))

		switch report.Type {
		case IPTS_REPORT_TYPE_TOUCH_HEATMAP_DIM:
			height, err := ipts.Protocol.ReadByte()
			if err != nil {
				break
			}

			width, err := ipts.Protocol.ReadByte()
			if err != nil {
				break
			}

			hm.Height = int(height)
			hm.Width = int(width)

			err = ipts.Protocol.Skip(6)
			break
		case IPTS_REPORT_TYPE_TOUCH_HEATMAP:
			hmb, ok := heatmapBufferCache[report.Size]
			if !ok {
				hmb = make([]byte, report.Size)
				heatmapBufferCache[report.Size] = hmb
			}
			hm.Data = hmb
			err = ipts.Protocol.Read(hmb)
			break
		default:
			// ignored
			err = ipts.Protocol.Skip(uint32(report.Size))
			break
		}

		if err != nil {
			return nil
		}
	}

	if len(hm.Data) == 0 {
		return nil
	}

	if len(hm.Data) != hm.Width*hm.Height {
		return nil
	}

	for i := 0; i < len(hm.Data); i++ {
		hm.Data[i] = 255 - hm.Data[i]
	}

	err := IptsTouchHandleHeatmap(ipts, hm)
	if err != nil {
		return err
	}

	return nil
}
