package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/0xThiebaut/PCAPeek/application"
	"github.com/0xThiebaut/PCAPeek/application/reverse"
	"github.com/0xThiebaut/PCAPeek/application/rfb"
	"github.com/0xThiebaut/PCAPeek/output/files"
	"github.com/0xThiebaut/PCAPeek/output/files/binary"
	"github.com/0xThiebaut/PCAPeek/output/files/fork"
	"github.com/0xThiebaut/PCAPeek/output/media"
	mfork "github.com/0xThiebaut/PCAPeek/output/media/fork"
	"github.com/0xThiebaut/PCAPeek/output/media/jpeg"
	"github.com/0xThiebaut/PCAPeek/output/media/mjpeg"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/tcpassembly"
	"github.com/spf13/cobra"
)

var (
	// PCAP settings
	bpf string
	// JPEG output
	useJpeg     = false
	jpegDir     = `./`
	jpegQuality = 100
	jpegFps     = 0
	// MJPEG
	useMjpeg     = false
	mjpegDir     = `./`
	mjpegQuality = 100
	mjpegFps     = 10
	// Files
	useFiles = false
	filesDir = `./`
)

var rootCmd = &cobra.Command{
	Use:   "PCAPeek PCAP [PCAP ...]",
	Short: "PCAPeek peeks into PCAPs",
	Long:  `PCAPeek is a tool to peek into PCAPs. It doesn't do much besides acting as a proof of concept to reconstruct reverse VNC traffic.`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Trim any BPF spaces
		bpf = strings.TrimSpace(bpf)

		// Prepare the media
		var mo []media.Factory
		if useJpeg {
			mo = append(mo, jpeg.New(jpegDir, jpegFps, jpegQuality))
		}
		if useMjpeg {
			mo = append(mo, mjpeg.New(mjpegDir, mjpegFps, mjpegQuality))
		}
		// Prepare the files
		var fo []files.Factory
		if useFiles {
			fo = append(fo, binary.New(filesDir))
		}
		// TODO: Abstract VNC extraction
		rrfb := reverse.New(rfb.New(mfork.New(mo...), fork.New(fo...)))
		pool := tcpassembly.NewStreamPool(application.NewApplicationStreamFactory(false, rrfb))
		assembler := tcpassembly.NewAssembler(pool)

		// Loop the PCAPs
		for _, file := range args {
			// Open the file
			handle, err := pcap.OpenOffline(file)
			if err != nil {
				return err
			}
			// Apply any BPF filters
			if len(bpf) > 0 {
				if err = handle.SetBPFFilter(bpf); err != nil {
					return err
				}
			}
			// Create the PCAP source
			decoder := gopacket.DecodersByLayerName[handle.LinkType().String()]
			source := gopacket.NewPacketSource(handle, decoder)
			source.Lazy = true
			source.NoCopy = true
			// Pipe the PCAPs
			for packet := range source.Packets() {
				if packet.NetworkLayer() == nil || packet.TransportLayer() == nil || packet.TransportLayer().LayerType() != layers.LayerTypeTCP {
					continue
				}
				tcp, ok := packet.TransportLayer().(*layers.TCP)
				if !ok {
					continue
				}
				// Set factory time here for each packet allowing on confirmed read to get the current timestamp or on first read to set the timestamp
				assembler.AssembleWithTimestamp(packet.NetworkLayer().NetworkFlow(), tcp, packet.Metadata().Timestamp)
			}
		}

		return nil
	},
}

func main() {
	rootCmd.PersistentFlags().StringVar(&bpf, `filter`, bpf, `A BPF filter to apply on the PCAPs`)
	rootCmd.PersistentFlags().BoolVar(&useJpeg, `jpeg`, useJpeg, `Output JPEG frames`)
	rootCmd.PersistentFlags().StringVar(&jpegDir, `jpeg-dir`, jpegDir, `The output directory for the JPEG frames`)
	rootCmd.PersistentFlags().IntVar(&jpegQuality, `jpeg-quality`, jpegQuality, `The JPEG frame quality percentage`)
	rootCmd.PersistentFlags().IntVar(&jpegFps, `jpeg-fps`, jpegFps, `The number of JPEG frames to output per second (default 0, outputs all frames)`)
	rootCmd.PersistentFlags().BoolVar(&useMjpeg, `mjpeg`, useMjpeg, `Output MJPEG videos`)
	rootCmd.PersistentFlags().StringVar(&mjpegDir, `mjpeg-dir`, mjpegDir, `The output directory for the MJPEG videos`)
	rootCmd.PersistentFlags().IntVar(&mjpegQuality, `mjpeg-quality`, mjpegQuality, `The MJPEG video quality percentage`)
	rootCmd.PersistentFlags().IntVar(&mjpegFps, `mjpeg-fps`, mjpegFps, `The number of MJPEG frames to output per second`)
	rootCmd.PersistentFlags().BoolVar(&useFiles, `files`, useFiles, `Output clipboard files`)
	rootCmd.PersistentFlags().StringVar(&filesDir, `files-dir`, filesDir, `The output directory for the clipboard files`)

	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
