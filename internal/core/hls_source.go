package core

import (
	"context"
	"time"

	"github.com/aler9/gortsplib/v2/pkg/format"
	"github.com/aler9/gortsplib/v2/pkg/media"

	"github.com/aler9/rtsp-simple-server/internal/conf"
	"github.com/aler9/rtsp-simple-server/internal/formatprocessor"
	"github.com/aler9/rtsp-simple-server/internal/hls"
	"github.com/aler9/rtsp-simple-server/internal/logger"
)

type hlsSourceParent interface {
	log(logger.Level, string, ...interface{})
	sourceStaticImplSetReady(req pathSourceStaticSetReadyReq) pathSourceStaticSetReadyRes
	sourceStaticImplSetNotReady(req pathSourceStaticSetNotReadyReq)
}

type hlsSource struct {
	parent hlsSourceParent
}

func newHLSSource(
	parent hlsSourceParent,
) *hlsSource {
	return &hlsSource{
		parent: parent,
	}
}

func (s *hlsSource) Log(level logger.Level, format string, args ...interface{}) {
	s.parent.log(level, "[hls source] "+format, args...)
}

// run implements sourceStaticImpl.
func (s *hlsSource) run(ctx context.Context, cnf *conf.PathConf, reloadConf chan *conf.PathConf) error {
	var stream *stream

	defer func() {
		if stream != nil {
			s.parent.sourceStaticImplSetNotReady(pathSourceStaticSetNotReadyReq{})
		}
	}()

	c, err := hls.NewClient(
		cnf.Source,
		cnf.SourceFingerprint,
		s,
	)
	if err != nil {
		return err
	}

	c.OnTracks(func(tracks []format.Format) error {
		var medias media.Medias

		for _, track := range tracks {
			medi := &media.Media{
				Formats: []format.Format{track},
			}
			medias = append(medias, medi)
			cformat := track

			switch track.(type) {
			case *format.H264:
				medi.Type = media.TypeVideo

				c.OnData(track, func(pts time.Duration, dat interface{}) {
					err := stream.writeData(medi, cformat, &formatprocessor.DataH264{
						PTS: pts,
						AU:  dat.([][]byte),
						NTP: time.Now(),
					})
					if err != nil {
						s.Log(logger.Warn, "%v", err)
					}
				})

			case *format.H265:
				medi.Type = media.TypeVideo

				c.OnData(track, func(pts time.Duration, dat interface{}) {
					err := stream.writeData(medi, cformat, &formatprocessor.DataH265{
						PTS: pts,
						AU:  dat.([][]byte),
						NTP: time.Now(),
					})
					if err != nil {
						s.Log(logger.Warn, "%v", err)
					}
				})

			case *format.MPEG4Audio:
				medi.Type = media.TypeAudio

				c.OnData(track, func(pts time.Duration, dat interface{}) {
					err := stream.writeData(medi, cformat, &formatprocessor.DataMPEG4Audio{
						PTS: pts,
						AUs: [][]byte{dat.([]byte)},
						NTP: time.Now(),
					})
					if err != nil {
						s.Log(logger.Warn, "%v", err)
					}
				})

			case *format.Opus:
				medi.Type = media.TypeAudio

				c.OnData(track, func(pts time.Duration, dat interface{}) {
					err := stream.writeData(medi, cformat, &formatprocessor.DataOpus{
						PTS:   pts,
						Frame: dat.([]byte),
						NTP:   time.Now(),
					})
					if err != nil {
						s.Log(logger.Warn, "%v", err)
					}
				})
			}
		}

		res := s.parent.sourceStaticImplSetReady(pathSourceStaticSetReadyReq{
			medias:             medias,
			generateRTPPackets: true,
		})
		if res.err != nil {
			return res.err
		}

		s.Log(logger.Info, "ready: %s", sourceMediaInfo(medias))
		stream = res.stream

		return nil
	})

	c.Start()

	for {
		select {
		case err := <-c.Wait():
			return err

		case <-reloadConf:

		case <-ctx.Done():
			c.Close()
			<-c.Wait()
			return nil
		}
	}
}

// apiSourceDescribe implements sourceStaticImpl.
func (*hlsSource) apiSourceDescribe() interface{} {
	return struct {
		Type string `json:"type"`
	}{"hlsSource"}
}
