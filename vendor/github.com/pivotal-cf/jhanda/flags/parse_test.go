package flags_test

import (
	"time"

	"github.com/pivotal-cf/jhanda/flags"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Parse", func() {
	Context("boolean flags", func() {
		It("parses short name flags", func() {
			var set struct {
				First  bool `short:"1"`
				Second bool `short:"2"`
			}
			args, err := flags.Parse(&set, []string{"-1", "command"})
			Expect(err).NotTo(HaveOccurred())
			Expect(args).To(Equal([]string{"command"}))

			Expect(set.First).To(BeTrue())
			Expect(set.Second).To(BeFalse())
		})

		It("parses long name flags", func() {
			var set struct {
				First  bool `long:"first"`
				Second bool `long:"second"`
			}
			args, err := flags.Parse(&set, []string{"--second", "command"})
			Expect(err).NotTo(HaveOccurred())
			Expect(args).To(Equal([]string{"command"}))

			Expect(set.First).To(BeFalse())
			Expect(set.Second).To(BeTrue())
		})

		It("allows for setting a default value", func() {
			var set struct {
				First  bool `long:"first" default:"true"`
				Second bool `long:"second"`
			}
			args, err := flags.Parse(&set, []string{"--second", "command"})
			Expect(err).NotTo(HaveOccurred())
			Expect(args).To(Equal([]string{"command"}))

			Expect(set.First).To(BeTrue())
			Expect(set.Second).To(BeTrue())
		})

		Context("when the default value is unparsable", func() {
			It("returns an error", func() {
				var set struct {
					First  bool `long:"first" default:"banana"`
					Second bool `long:"second"`
				}
				_, err := flags.Parse(&set, []string{"--second", "command"})
				Expect(err).To(MatchError(ContainSubstring("could not parse bool default value \"banana\"")))
			})
		})
	})

	Context("slice flags", func() {
		It("parses short name flags", func() {
			var set struct {
				First  flags.StringSlice `short:"1"`
				Second flags.StringSlice `short:"2"`
			}
			args, err := flags.Parse(&set, []string{"-1", "test", "-1", "another-test", "command"})
			Expect(err).NotTo(HaveOccurred())
			Expect(args).To(Equal([]string{"command"}))

			Expect(set.First).To(ConsistOf(flags.StringSlice{"test", "another-test"}))
			Expect(set.Second).To(BeEmpty())
		})

		It("parses long name flags", func() {
			var set struct {
				First  flags.StringSlice `long:"first"`
				Second flags.StringSlice `long:"second"`
			}
			args, err := flags.Parse(&set, []string{"--second", "test", "--second", "different-test", "command"})
			Expect(err).NotTo(HaveOccurred())
			Expect(args).To(Equal([]string{"command"}))

			Expect(set.First).To(BeEmpty())
			Expect(set.Second).To(ConsistOf(flags.StringSlice{"test", "different-test"}))
		})

		It("allows for setting a default value", func() {
			var set struct {
				First  flags.StringSlice `long:"first" default:"yes,no"`
				Second flags.StringSlice `long:"second"`
			}
			args, err := flags.Parse(&set, []string{"--second", "what", "command"})
			Expect(err).NotTo(HaveOccurred())
			Expect(args).To(Equal([]string{"command"}))

			Expect(set.First).To(ConsistOf(flags.StringSlice{"yes", "no"}))
			Expect(set.Second).To(ConsistOf(flags.StringSlice{"what"}))
		})
	})

	Context("float64 flags", func() {
		It("parses short name flags", func() {
			var set struct {
				First  float64 `short:"1"`
				Second float64 `short:"2"`
			}
			args, err := flags.Parse(&set, []string{"-1", "12.3", "command"})
			Expect(err).NotTo(HaveOccurred())
			Expect(args).To(Equal([]string{"command"}))

			Expect(set.First).To(Equal(12.3))
			Expect(set.Second).To(Equal(0.0))
		})

		It("parses long name flags", func() {
			var set struct {
				First  float64 `long:"first"`
				Second float64 `long:"second"`
			}
			args, err := flags.Parse(&set, []string{"--second", "45.6", "command"})
			Expect(err).NotTo(HaveOccurred())
			Expect(args).To(Equal([]string{"command"}))

			Expect(set.First).To(Equal(0.0))
			Expect(set.Second).To(Equal(45.6))
		})

		It("allows for setting a default value", func() {
			var set struct {
				First  float64 `long:"first" default:"78.9"`
				Second float64 `long:"second"`
			}
			args, err := flags.Parse(&set, []string{"--second", "12.3", "command"})
			Expect(err).NotTo(HaveOccurred())
			Expect(args).To(Equal([]string{"command"}))

			Expect(set.First).To(Equal(78.9))
			Expect(set.Second).To(Equal(12.3))
		})

		Context("when the default value is unparsable", func() {
			It("returns an error", func() {
				var set struct {
					First  float64 `long:"first" default:"banana"`
					Second float64 `long:"second"`
				}
				_, err := flags.Parse(&set, []string{"--second", "command"})
				Expect(err).To(MatchError(ContainSubstring("could not parse float64 default value \"banana\"")))
			})
		})
	})

	Context("int64 flags", func() {
		It("parses short name flags", func() {
			var set struct {
				First  int64 `short:"1"`
				Second int64 `short:"2"`
			}
			args, err := flags.Parse(&set, []string{"-1", "345", "command"})
			Expect(err).NotTo(HaveOccurred())
			Expect(args).To(Equal([]string{"command"}))

			Expect(set.First).To(Equal(int64(345)))
			Expect(set.Second).To(Equal(int64(0)))
		})

		It("parses long name flags", func() {
			var set struct {
				First  int64 `long:"first"`
				Second int64 `long:"second"`
			}
			args, err := flags.Parse(&set, []string{"--second", "789", "command"})
			Expect(err).NotTo(HaveOccurred())
			Expect(args).To(Equal([]string{"command"}))

			Expect(set.First).To(Equal(int64(0)))
			Expect(set.Second).To(Equal(int64(789)))
		})

		It("allows for setting a default value", func() {
			var set struct {
				First  int64 `long:"first" default:"123"`
				Second int64 `long:"second"`
			}
			args, err := flags.Parse(&set, []string{"--second", "999", "command"})
			Expect(err).NotTo(HaveOccurred())
			Expect(args).To(Equal([]string{"command"}))

			Expect(set.First).To(Equal(int64(123)))
			Expect(set.Second).To(Equal(int64(999)))
		})

		Context("when the default value is unparsable", func() {
			It("returns an error", func() {
				var set struct {
					First  int64 `long:"first" default:"banana"`
					Second int64 `long:"second"`
				}
				_, err := flags.Parse(&set, []string{"--second", "command"})
				Expect(err).To(MatchError(ContainSubstring("could not parse int64 default value \"banana\"")))
			})
		})
	})

	Context("duration flags", func() {
		It("parses short name flags", func() {
			var set struct {
				First  time.Duration `short:"1"`
				Second time.Duration `short:"2"`
			}
			args, err := flags.Parse(&set, []string{"-1", "1s", "command"})
			Expect(err).NotTo(HaveOccurred())
			Expect(args).To(Equal([]string{"command"}))

			Expect(set.First).To(Equal(1 * time.Second))
			Expect(set.Second).To(Equal(time.Duration(0)))
		})

		It("parses long name flags", func() {
			var set struct {
				First  time.Duration `long:"first"`
				Second time.Duration `long:"second"`
			}
			args, err := flags.Parse(&set, []string{"--second", "45m", "command"})
			Expect(err).NotTo(HaveOccurred())
			Expect(args).To(Equal([]string{"command"}))

			Expect(set.First).To(Equal(time.Duration(0)))
			Expect(set.Second).To(Equal(45 * time.Minute))
		})

		It("allows for setting a default value", func() {
			var set struct {
				First  time.Duration `long:"first" default:"23ms"`
				Second time.Duration `long:"second"`
			}
			args, err := flags.Parse(&set, []string{"--second", "42h", "command"})
			Expect(err).NotTo(HaveOccurred())
			Expect(args).To(Equal([]string{"command"}))

			Expect(set.First).To(Equal(23 * time.Millisecond))
			Expect(set.Second).To(Equal(42 * time.Hour))
		})

		Context("when the default value is unparsable", func() {
			It("returns an error", func() {
				var set struct {
					First  time.Duration `long:"first" default:"banana"`
					Second time.Duration `long:"second"`
				}
				_, err := flags.Parse(&set, []string{"--second", "command"})
				Expect(err).To(MatchError(ContainSubstring("could not parse duration default value \"banana\"")))
			})
		})
	})

	Context("int flags", func() {
		It("parses short name flags", func() {
			var set struct {
				First  int `short:"1"`
				Second int `short:"2"`
			}
			args, err := flags.Parse(&set, []string{"-1", "123", "command"})
			Expect(err).NotTo(HaveOccurred())
			Expect(args).To(Equal([]string{"command"}))

			Expect(set.First).To(Equal(123))
			Expect(set.Second).To(Equal(0))
		})

		It("parses long name flags", func() {
			var set struct {
				First  int `long:"first"`
				Second int `long:"second"`
			}
			args, err := flags.Parse(&set, []string{"--second", "456", "command"})
			Expect(err).NotTo(HaveOccurred())
			Expect(args).To(Equal([]string{"command"}))

			Expect(set.First).To(Equal(0))
			Expect(set.Second).To(Equal(456))
		})

		It("allows for setting a default value", func() {
			var set struct {
				First  int `long:"first" default:"234"`
				Second int `long:"second"`
			}
			args, err := flags.Parse(&set, []string{"--second", "420", "command"})
			Expect(err).NotTo(HaveOccurred())
			Expect(args).To(Equal([]string{"command"}))

			Expect(set.First).To(Equal(234))
			Expect(set.Second).To(Equal(420))
		})

		Context("when the default value is unparsable", func() {
			It("returns an error", func() {
				var set struct {
					First  int `long:"first" default:"banana"`
					Second int `long:"second"`
				}
				_, err := flags.Parse(&set, []string{"--second", "command"})
				Expect(err).To(MatchError(ContainSubstring("could not parse int default value \"banana\"")))
			})
		})
	})

	Context("string flags", func() {
		It("parses short name flags", func() {
			var set struct {
				First  string `short:"1"`
				Second string `short:"2"`
			}
			args, err := flags.Parse(&set, []string{"-1", "hello", "command"})
			Expect(err).NotTo(HaveOccurred())
			Expect(args).To(Equal([]string{"command"}))

			Expect(set.First).To(Equal("hello"))
			Expect(set.Second).To(Equal(""))
		})

		It("parses long name flags", func() {
			var set struct {
				First  string `long:"first"`
				Second string `long:"second"`
			}
			args, err := flags.Parse(&set, []string{"--second", "goodbye", "command"})
			Expect(err).NotTo(HaveOccurred())
			Expect(args).To(Equal([]string{"command"}))

			Expect(set.First).To(Equal(""))
			Expect(set.Second).To(Equal("goodbye"))
		})

		It("allows for setting a default value", func() {
			var set struct {
				First  string `long:"first" default:"default"`
				Second string `long:"second"`
			}
			args, err := flags.Parse(&set, []string{"--second", "custom", "command"})
			Expect(err).NotTo(HaveOccurred())
			Expect(args).To(Equal([]string{"command"}))

			Expect(set.First).To(Equal("default"))
			Expect(set.Second).To(Equal("custom"))
		})
	})

	Context("uint64 flags", func() {
		It("parses short name flags", func() {
			var set struct {
				First  uint64 `short:"1"`
				Second uint64 `short:"2"`
			}
			args, err := flags.Parse(&set, []string{"-1", "123", "command"})
			Expect(err).NotTo(HaveOccurred())
			Expect(args).To(Equal([]string{"command"}))

			Expect(set.First).To(Equal(uint64(123)))
			Expect(set.Second).To(Equal(uint64(0)))
		})

		It("parses long name flags", func() {
			var set struct {
				First  uint64 `long:"first"`
				Second uint64 `long:"second"`
			}
			args, err := flags.Parse(&set, []string{"--second", "456", "command"})
			Expect(err).NotTo(HaveOccurred())
			Expect(args).To(Equal([]string{"command"}))

			Expect(set.First).To(Equal(uint64(0)))
			Expect(set.Second).To(Equal(uint64(456)))
		})

		It("allows for setting a default value", func() {
			var set struct {
				First  uint64 `long:"first" default:"234"`
				Second uint64 `long:"second"`
			}
			args, err := flags.Parse(&set, []string{"--second", "420", "command"})
			Expect(err).NotTo(HaveOccurred())
			Expect(args).To(Equal([]string{"command"}))

			Expect(set.First).To(Equal(uint64(234)))
			Expect(set.Second).To(Equal(uint64(420)))
		})

		Context("when the default value is unparsable", func() {
			It("returns an error", func() {
				var set struct {
					First  uint64 `long:"first" default:"banana"`
					Second uint64 `long:"second"`
				}
				_, err := flags.Parse(&set, []string{"--second", "command"})
				Expect(err).To(MatchError(ContainSubstring("could not parse uint64 default value \"banana\"")))
			})
		})
	})

	Context("uint flags", func() {
		It("parses short name flags", func() {
			var set struct {
				First  uint `short:"1"`
				Second uint `short:"2"`
			}
			args, err := flags.Parse(&set, []string{"-1", "123", "command"})
			Expect(err).NotTo(HaveOccurred())
			Expect(args).To(Equal([]string{"command"}))

			Expect(set.First).To(Equal(uint(123)))
			Expect(set.Second).To(Equal(uint(0)))
		})

		It("parses long name flags", func() {
			var set struct {
				First  uint `long:"first"`
				Second uint `long:"second"`
			}
			args, err := flags.Parse(&set, []string{"--second", "456", "command"})
			Expect(err).NotTo(HaveOccurred())
			Expect(args).To(Equal([]string{"command"}))

			Expect(set.First).To(Equal(uint(0)))
			Expect(set.Second).To(Equal(uint(456)))
		})

		It("allows for setting a default value", func() {
			var set struct {
				First  uint `long:"first" default:"234"`
				Second uint `long:"second"`
			}
			args, err := flags.Parse(&set, []string{"--second", "420", "command"})
			Expect(err).NotTo(HaveOccurred())
			Expect(args).To(Equal([]string{"command"}))

			Expect(set.First).To(Equal(uint(234)))
			Expect(set.Second).To(Equal(uint(420)))
		})

		Context("when the default value is unparsable", func() {
			It("returns an error", func() {
				var set struct {
					First  uint `long:"first" default:"banana"`
					Second uint `long:"second"`
				}
				_, err := flags.Parse(&set, []string{"--second", "command"})
				Expect(err).To(MatchError(ContainSubstring("could not parse uint default value \"banana\"")))
			})
		})
	})

	Context("failure cases", func() {
		Context("when a non-pointer is passed as the receiver", func() {
			It("returns an error", func() {
				var set struct {
					First  bool `long:"first"`
					Second bool `long:"second"`
				}
				_, err := flags.Parse(set, []string{"--second", "command"})
				Expect(err).To(MatchError("unexpected non-pointer type struct for flag receiver"))
			})
		})

		Context("when the receiver does not point to a struct", func() {
			It("returns an error", func() {
				var notAStruct int
				_, err := flags.Parse(&notAStruct, []string{"--second", "command"})
				Expect(err).To(MatchError("unexpected pointer to non-struct type int"))
			})
		})

		Context("when the receiver has an unsupported flag type", func() {
			It("returns an error", func() {
				var set struct {
					Unsupported func()
				}
				_, err := flags.Parse(&set, []string{"--second", "command"})
				Expect(err).To(MatchError("unexpected flag receiver field type func"))
			})
		})

		Context("when a flag is unknown", func() {
			It("returns an error", func() {
				var set struct {
					First  bool `long:"first"`
					Second bool `long:"second"`
				}
				_, err := flags.Parse(&set, []string{"--third", "command"})
				Expect(err).To(MatchError("flag provided but not defined: -third"))
			})
		})
	})
})
