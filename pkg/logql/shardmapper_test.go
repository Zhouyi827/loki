package logql

import (
	"testing"
	"time"

	"github.com/cortexproject/cortex/pkg/querier/astmapper"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/stretchr/testify/require"
)

func TestStringer(t *testing.T) {
	for _, tc := range []struct {
		in  Expr
		out string
	}{
		{
			in: &ConcatLogSelectorExpr{
				LogSelectorExpr: DownstreamLogSelectorExpr{
					shard: &astmapper.ShardAnnotation{
						Shard: 0,
						Of:    2,
					},
					LogSelectorExpr: &matchersExpr{
						matchers: []*labels.Matcher{
							mustNewMatcher(labels.MatchEqual, "foo", "bar"),
						},
					},
				},
				next: &ConcatLogSelectorExpr{
					LogSelectorExpr: DownstreamLogSelectorExpr{
						shard: &astmapper.ShardAnnotation{
							Shard: 1,
							Of:    2,
						},
						LogSelectorExpr: &matchersExpr{
							matchers: []*labels.Matcher{
								mustNewMatcher(labels.MatchEqual, "foo", "bar"),
							},
						},
					},
					next: nil,
				},
			},
			out: `downstream<{foo="bar"}, shard=0_of_2> ++ downstream<{foo="bar"}, shard=1_of_2>`,
		},
	} {
		t.Run(tc.out, func(t *testing.T) {
			require.Equal(t, tc.out, tc.in.String())
		})
	}
}

func TestMapSampleExpr(t *testing.T) {
	m, err := NewShardMapper(2)
	require.Nil(t, err)

	for _, tc := range []struct {
		in  SampleExpr
		out SampleExpr
	}{
		{
			in: &rangeAggregationExpr{
				operation: OpTypeRate,
				left: &logRange{
					left: &matchersExpr{
						matchers: []*labels.Matcher{
							mustNewMatcher(labels.MatchEqual, "foo", "bar"),
						},
					},
					interval: time.Minute,
				},
			},
			out: &ConcatSampleExpr{
				SampleExpr: DownstreamSampleExpr{
					shard: &astmapper.ShardAnnotation{
						Shard: 0,
						Of:    2,
					},
					SampleExpr: &rangeAggregationExpr{
						operation: OpTypeRate,
						left: &logRange{
							left: &matchersExpr{
								matchers: []*labels.Matcher{
									mustNewMatcher(labels.MatchEqual, "foo", "bar"),
								},
							},
							interval: time.Minute,
						},
					},
				},
				next: &ConcatSampleExpr{
					SampleExpr: DownstreamSampleExpr{
						shard: &astmapper.ShardAnnotation{
							Shard: 1,
							Of:    2,
						},
						SampleExpr: &rangeAggregationExpr{
							operation: OpTypeRate,
							left: &logRange{
								left: &matchersExpr{
									matchers: []*labels.Matcher{
										mustNewMatcher(labels.MatchEqual, "foo", "bar"),
									},
								},
								interval: time.Minute,
							},
						},
					},
					next: nil,
				},
			},
		},
	} {
		t.Run(tc.in.String(), func(t *testing.T) {
			require.Equal(t, tc.out, m.mapSampleExpr(tc.in))
		})

	}
}

func TestMapping(t *testing.T) {
	m, err := NewShardMapper(2)
	require.Nil(t, err)

	for _, tc := range []struct {
		in   string
		expr Expr
		err  error
	}{
		{
			in: `{foo="bar"}`,
			expr: &ConcatLogSelectorExpr{
				LogSelectorExpr: DownstreamLogSelectorExpr{
					shard: &astmapper.ShardAnnotation{
						Shard: 0,
						Of:    2,
					},
					LogSelectorExpr: &matchersExpr{
						matchers: []*labels.Matcher{
							mustNewMatcher(labels.MatchEqual, "foo", "bar"),
						},
					},
				},
				next: &ConcatLogSelectorExpr{
					LogSelectorExpr: DownstreamLogSelectorExpr{
						shard: &astmapper.ShardAnnotation{
							Shard: 1,
							Of:    2,
						},
						LogSelectorExpr: &matchersExpr{
							matchers: []*labels.Matcher{
								mustNewMatcher(labels.MatchEqual, "foo", "bar"),
							},
						},
					},
					next: nil,
				},
			},
		},
		{
			in: `{foo="bar"} |= "error"`,
			expr: &ConcatLogSelectorExpr{
				LogSelectorExpr: DownstreamLogSelectorExpr{
					shard: &astmapper.ShardAnnotation{
						Shard: 0,
						Of:    2,
					},
					LogSelectorExpr: &filterExpr{
						match: "error",
						ty:    labels.MatchEqual,
						left: &matchersExpr{
							matchers: []*labels.Matcher{
								mustNewMatcher(labels.MatchEqual, "foo", "bar"),
							},
						},
					},
				},
				next: &ConcatLogSelectorExpr{
					LogSelectorExpr: DownstreamLogSelectorExpr{
						shard: &astmapper.ShardAnnotation{
							Shard: 1,
							Of:    2,
						},
						LogSelectorExpr: &filterExpr{
							match: "error",
							ty:    labels.MatchEqual,
							left: &matchersExpr{
								matchers: []*labels.Matcher{
									mustNewMatcher(labels.MatchEqual, "foo", "bar"),
								},
							},
						},
					},
					next: nil,
				},
			},
		},
		{
			in: `rate({foo="bar"}[5m])`,
			expr: &ConcatSampleExpr{
				SampleExpr: DownstreamSampleExpr{
					shard: &astmapper.ShardAnnotation{
						Shard: 0,
						Of:    2,
					},
					SampleExpr: &rangeAggregationExpr{
						operation: OpTypeRate,
						left: &logRange{
							left: &matchersExpr{
								matchers: []*labels.Matcher{
									mustNewMatcher(labels.MatchEqual, "foo", "bar"),
								},
							},
							interval: 5 * time.Minute,
						},
					},
				},
				next: &ConcatSampleExpr{
					SampleExpr: DownstreamSampleExpr{
						shard: &astmapper.ShardAnnotation{
							Shard: 1,
							Of:    2,
						},
						SampleExpr: &rangeAggregationExpr{
							operation: OpTypeRate,
							left: &logRange{
								left: &matchersExpr{
									matchers: []*labels.Matcher{
										mustNewMatcher(labels.MatchEqual, "foo", "bar"),
									},
								},
								interval: 5 * time.Minute,
							},
						},
					},
					next: nil,
				},
			},
		},
		{
			in: `count_over_time({foo="bar"}[5m])`,
			expr: &ConcatSampleExpr{
				SampleExpr: DownstreamSampleExpr{
					shard: &astmapper.ShardAnnotation{
						Shard: 0,
						Of:    2,
					},
					SampleExpr: &rangeAggregationExpr{
						operation: OpTypeCountOverTime,
						left: &logRange{
							left: &matchersExpr{
								matchers: []*labels.Matcher{
									mustNewMatcher(labels.MatchEqual, "foo", "bar"),
								},
							},
							interval: 5 * time.Minute,
						},
					},
				},
				next: &ConcatSampleExpr{
					SampleExpr: DownstreamSampleExpr{
						shard: &astmapper.ShardAnnotation{
							Shard: 1,
							Of:    2,
						},
						SampleExpr: &rangeAggregationExpr{
							operation: OpTypeCountOverTime,
							left: &logRange{
								left: &matchersExpr{
									matchers: []*labels.Matcher{
										mustNewMatcher(labels.MatchEqual, "foo", "bar"),
									},
								},
								interval: 5 * time.Minute,
							},
						},
					},
					next: nil,
				},
			},
		},
		{
			in: `sum(rate({foo="bar"}[5m]))`,
			expr: &vectorAggregationExpr{
				grouping:  &grouping{},
				operation: OpTypeSum,
				left: &ConcatSampleExpr{
					SampleExpr: DownstreamSampleExpr{
						shard: &astmapper.ShardAnnotation{
							Shard: 0,
							Of:    2,
						},
						SampleExpr: &vectorAggregationExpr{
							grouping:  &grouping{},
							operation: OpTypeSum,
							left: &rangeAggregationExpr{
								operation: OpTypeRate,
								left: &logRange{
									left: &matchersExpr{
										matchers: []*labels.Matcher{
											mustNewMatcher(labels.MatchEqual, "foo", "bar"),
										},
									},
									interval: 5 * time.Minute,
								},
							},
						},
					},
					next: &ConcatSampleExpr{
						SampleExpr: DownstreamSampleExpr{
							shard: &astmapper.ShardAnnotation{
								Shard: 1,
								Of:    2,
							},
							SampleExpr: &vectorAggregationExpr{
								grouping:  &grouping{},
								operation: OpTypeSum,
								left: &rangeAggregationExpr{
									operation: OpTypeRate,
									left: &logRange{
										left: &matchersExpr{
											matchers: []*labels.Matcher{
												mustNewMatcher(labels.MatchEqual, "foo", "bar"),
											},
										},
										interval: 5 * time.Minute,
									},
								},
							},
						},
						next: nil,
					},
				},
			},
		},
		{
			in: `topk(3, rate({foo="bar"}[5m]))`,
			expr: &vectorAggregationExpr{
				grouping:  &grouping{},
				params:    3,
				operation: OpTypeTopK,
				left: &ConcatSampleExpr{
					SampleExpr: DownstreamSampleExpr{
						shard: &astmapper.ShardAnnotation{
							Shard: 0,
							Of:    2,
						},
						SampleExpr: &rangeAggregationExpr{
							operation: OpTypeRate,
							left: &logRange{
								left: &matchersExpr{
									matchers: []*labels.Matcher{
										mustNewMatcher(labels.MatchEqual, "foo", "bar"),
									},
								},
								interval: 5 * time.Minute,
							},
						},
					},
					next: &ConcatSampleExpr{
						SampleExpr: DownstreamSampleExpr{
							shard: &astmapper.ShardAnnotation{
								Shard: 1,
								Of:    2,
							},
							SampleExpr: &rangeAggregationExpr{
								operation: OpTypeRate,
								left: &logRange{
									left: &matchersExpr{
										matchers: []*labels.Matcher{
											mustNewMatcher(labels.MatchEqual, "foo", "bar"),
										},
									},
									interval: 5 * time.Minute,
								},
							},
						},
						next: nil,
					},
				},
			},
		},
		{
			in: `max without (env) (rate({foo="bar"}[5m]))`,
			expr: &vectorAggregationExpr{
				grouping: &grouping{
					without: true,
					groups:  []string{"env"},
				},
				operation: OpTypeMax,
				left: &ConcatSampleExpr{
					SampleExpr: DownstreamSampleExpr{
						shard: &astmapper.ShardAnnotation{
							Shard: 0,
							Of:    2,
						},
						SampleExpr: &rangeAggregationExpr{
							operation: OpTypeRate,
							left: &logRange{
								left: &matchersExpr{
									matchers: []*labels.Matcher{
										mustNewMatcher(labels.MatchEqual, "foo", "bar"),
									},
								},
								interval: 5 * time.Minute,
							},
						},
					},
					next: &ConcatSampleExpr{
						SampleExpr: DownstreamSampleExpr{
							shard: &astmapper.ShardAnnotation{
								Shard: 1,
								Of:    2,
							},
							SampleExpr: &rangeAggregationExpr{
								operation: OpTypeRate,
								left: &logRange{
									left: &matchersExpr{
										matchers: []*labels.Matcher{
											mustNewMatcher(labels.MatchEqual, "foo", "bar"),
										},
									},
									interval: 5 * time.Minute,
								},
							},
						},
						next: nil,
					},
				},
			},
		},
		{
			in: `count(rate({foo="bar"}[5m]))`,
			expr: &vectorAggregationExpr{
				operation: OpTypeSum,
				grouping:  &grouping{},
				left: &ConcatSampleExpr{
					SampleExpr: DownstreamSampleExpr{
						shard: &astmapper.ShardAnnotation{
							Shard: 0,
							Of:    2,
						},
						SampleExpr: &vectorAggregationExpr{
							grouping:  &grouping{},
							operation: OpTypeCount,
							left: &rangeAggregationExpr{
								operation: OpTypeRate,
								left: &logRange{
									left: &matchersExpr{
										matchers: []*labels.Matcher{
											mustNewMatcher(labels.MatchEqual, "foo", "bar"),
										},
									},
									interval: 5 * time.Minute,
								},
							},
						},
					},
					next: &ConcatSampleExpr{
						SampleExpr: DownstreamSampleExpr{
							shard: &astmapper.ShardAnnotation{
								Shard: 1,
								Of:    2,
							},
							SampleExpr: &vectorAggregationExpr{
								grouping:  &grouping{},
								operation: OpTypeCount,
								left: &rangeAggregationExpr{
									operation: OpTypeRate,
									left: &logRange{
										left: &matchersExpr{
											matchers: []*labels.Matcher{
												mustNewMatcher(labels.MatchEqual, "foo", "bar"),
											},
										},
										interval: 5 * time.Minute,
									},
								},
							},
						},
						next: nil,
					},
				},
			},
		},
		{
			in: `avg(rate({foo="bar"}[5m]))`,
			expr: &binOpExpr{
				op: OpTypeDiv,
				SampleExpr: &vectorAggregationExpr{
					grouping:  &grouping{},
					operation: OpTypeSum,
					left: &ConcatSampleExpr{
						SampleExpr: DownstreamSampleExpr{
							shard: &astmapper.ShardAnnotation{
								Shard: 0,
								Of:    2,
							},
							SampleExpr: &vectorAggregationExpr{
								grouping:  &grouping{},
								operation: OpTypeSum,
								left: &rangeAggregationExpr{
									operation: OpTypeRate,
									left: &logRange{
										left: &matchersExpr{
											matchers: []*labels.Matcher{
												mustNewMatcher(labels.MatchEqual, "foo", "bar"),
											},
										},
										interval: 5 * time.Minute,
									},
								},
							},
						},
						next: &ConcatSampleExpr{
							SampleExpr: DownstreamSampleExpr{
								shard: &astmapper.ShardAnnotation{
									Shard: 1,
									Of:    2,
								},
								SampleExpr: &vectorAggregationExpr{
									grouping:  &grouping{},
									operation: OpTypeSum,
									left: &rangeAggregationExpr{
										operation: OpTypeRate,
										left: &logRange{
											left: &matchersExpr{
												matchers: []*labels.Matcher{
													mustNewMatcher(labels.MatchEqual, "foo", "bar"),
												},
											},
											interval: 5 * time.Minute,
										},
									},
								},
							},
							next: nil,
						},
					},
				},
				RHS: &vectorAggregationExpr{
					operation: OpTypeSum,
					grouping:  &grouping{},
					left: &ConcatSampleExpr{
						SampleExpr: DownstreamSampleExpr{
							shard: &astmapper.ShardAnnotation{
								Shard: 0,
								Of:    2,
							},
							SampleExpr: &vectorAggregationExpr{
								grouping:  &grouping{},
								operation: OpTypeCount,
								left: &rangeAggregationExpr{
									operation: OpTypeRate,
									left: &logRange{
										left: &matchersExpr{
											matchers: []*labels.Matcher{
												mustNewMatcher(labels.MatchEqual, "foo", "bar"),
											},
										},
										interval: 5 * time.Minute,
									},
								},
							},
						},
						next: &ConcatSampleExpr{
							SampleExpr: DownstreamSampleExpr{
								shard: &astmapper.ShardAnnotation{
									Shard: 1,
									Of:    2,
								},
								SampleExpr: &vectorAggregationExpr{
									grouping:  &grouping{},
									operation: OpTypeCount,
									left: &rangeAggregationExpr{
										operation: OpTypeRate,
										left: &logRange{
											left: &matchersExpr{
												matchers: []*labels.Matcher{
													mustNewMatcher(labels.MatchEqual, "foo", "bar"),
												},
											},
											interval: 5 * time.Minute,
										},
									},
								},
							},
							next: nil,
						},
					},
				},
			},
		},
		{
			in: `1 + sum by (cluster) (rate({foo="bar"}[5m]))`,
			expr: &binOpExpr{
				op:         OpTypeAdd,
				SampleExpr: &literalExpr{1},
				RHS: &vectorAggregationExpr{
					grouping: &grouping{
						groups: []string{"cluster"},
					},
					operation: OpTypeSum,
					left: &ConcatSampleExpr{
						SampleExpr: DownstreamSampleExpr{
							shard: &astmapper.ShardAnnotation{
								Shard: 0,
								Of:    2,
							},
							SampleExpr: &vectorAggregationExpr{
								grouping: &grouping{
									groups: []string{"cluster"},
								},
								operation: OpTypeSum,
								left: &rangeAggregationExpr{
									operation: OpTypeRate,
									left: &logRange{
										left: &matchersExpr{
											matchers: []*labels.Matcher{
												mustNewMatcher(labels.MatchEqual, "foo", "bar"),
											},
										},
										interval: 5 * time.Minute,
									},
								},
							},
						},
						next: &ConcatSampleExpr{
							SampleExpr: DownstreamSampleExpr{
								shard: &astmapper.ShardAnnotation{
									Shard: 1,
									Of:    2,
								},
								SampleExpr: &vectorAggregationExpr{
									grouping: &grouping{
										groups: []string{"cluster"},
									},
									operation: OpTypeSum,
									left: &rangeAggregationExpr{
										operation: OpTypeRate,
										left: &logRange{
											left: &matchersExpr{
												matchers: []*labels.Matcher{
													mustNewMatcher(labels.MatchEqual, "foo", "bar"),
												},
											},
											interval: 5 * time.Minute,
										},
									},
								},
							},
							next: nil,
						},
					},
				},
			},
		},
		// sum(max) should not shard the maxes
		{
			in: `sum(max(rate({foo="bar"}[5m])))`,
			expr: &vectorAggregationExpr{
				grouping:  &grouping{},
				operation: OpTypeSum,
				left: &vectorAggregationExpr{
					grouping:  &grouping{},
					operation: OpTypeMax,
					left: &ConcatSampleExpr{
						SampleExpr: DownstreamSampleExpr{
							shard: &astmapper.ShardAnnotation{
								Shard: 0,
								Of:    2,
							},
							SampleExpr: &rangeAggregationExpr{
								operation: OpTypeRate,
								left: &logRange{
									left: &matchersExpr{
										matchers: []*labels.Matcher{
											mustNewMatcher(labels.MatchEqual, "foo", "bar"),
										},
									},
									interval: 5 * time.Minute,
								},
							},
						},
						next: &ConcatSampleExpr{
							SampleExpr: DownstreamSampleExpr{
								shard: &astmapper.ShardAnnotation{
									Shard: 1,
									Of:    2,
								},
								SampleExpr: &rangeAggregationExpr{
									operation: OpTypeRate,
									left: &logRange{
										left: &matchersExpr{
											matchers: []*labels.Matcher{
												mustNewMatcher(labels.MatchEqual, "foo", "bar"),
											},
										},
										interval: 5 * time.Minute,
									},
								},
							},
							next: nil,
						},
					},
				},
			},
		},
		// max(count) should shard the count, but not the max
		{
			in: `max(count(rate({foo="bar"}[5m])))`,
			expr: &vectorAggregationExpr{
				grouping:  &grouping{},
				operation: OpTypeMax,
				left: &vectorAggregationExpr{
					operation: OpTypeSum,
					grouping:  &grouping{},
					left: &ConcatSampleExpr{
						SampleExpr: DownstreamSampleExpr{
							shard: &astmapper.ShardAnnotation{
								Shard: 0,
								Of:    2,
							},
							SampleExpr: &vectorAggregationExpr{
								grouping:  &grouping{},
								operation: OpTypeCount,
								left: &rangeAggregationExpr{
									operation: OpTypeRate,
									left: &logRange{
										left: &matchersExpr{
											matchers: []*labels.Matcher{
												mustNewMatcher(labels.MatchEqual, "foo", "bar"),
											},
										},
										interval: 5 * time.Minute,
									},
								},
							},
						},
						next: &ConcatSampleExpr{
							SampleExpr: DownstreamSampleExpr{
								shard: &astmapper.ShardAnnotation{
									Shard: 1,
									Of:    2,
								},
								SampleExpr: &vectorAggregationExpr{
									grouping:  &grouping{},
									operation: OpTypeCount,
									left: &rangeAggregationExpr{
										operation: OpTypeRate,
										left: &logRange{
											left: &matchersExpr{
												matchers: []*labels.Matcher{
													mustNewMatcher(labels.MatchEqual, "foo", "bar"),
												},
											},
											interval: 5 * time.Minute,
										},
									},
								},
							},
							next: nil,
						},
					},
				},
			},
		},
		{
			in: `max(sum by (cluster) (rate({foo="bar"}[5m]))) / count(rate({foo="bar"}[5m]))`,
			expr: &binOpExpr{
				op: OpTypeDiv,
				SampleExpr: &vectorAggregationExpr{
					operation: OpTypeMax,
					grouping:  &grouping{},
					left: &vectorAggregationExpr{
						grouping: &grouping{
							groups: []string{"cluster"},
						},
						operation: OpTypeSum,
						left: &ConcatSampleExpr{
							SampleExpr: DownstreamSampleExpr{
								shard: &astmapper.ShardAnnotation{
									Shard: 0,
									Of:    2,
								},
								SampleExpr: &vectorAggregationExpr{
									grouping: &grouping{
										groups: []string{"cluster"},
									},
									operation: OpTypeSum,
									left: &rangeAggregationExpr{
										operation: OpTypeRate,
										left: &logRange{
											left: &matchersExpr{
												matchers: []*labels.Matcher{
													mustNewMatcher(labels.MatchEqual, "foo", "bar"),
												},
											},
											interval: 5 * time.Minute,
										},
									},
								},
							},
							next: &ConcatSampleExpr{
								SampleExpr: DownstreamSampleExpr{
									shard: &astmapper.ShardAnnotation{
										Shard: 1,
										Of:    2,
									},
									SampleExpr: &vectorAggregationExpr{
										grouping: &grouping{
											groups: []string{"cluster"},
										},
										operation: OpTypeSum,
										left: &rangeAggregationExpr{
											operation: OpTypeRate,
											left: &logRange{
												left: &matchersExpr{
													matchers: []*labels.Matcher{
														mustNewMatcher(labels.MatchEqual, "foo", "bar"),
													},
												},
												interval: 5 * time.Minute,
											},
										},
									},
								},
								next: nil,
							},
						},
					},
				},
				RHS: &vectorAggregationExpr{
					operation: OpTypeSum,
					grouping:  &grouping{},
					left: &ConcatSampleExpr{
						SampleExpr: DownstreamSampleExpr{
							shard: &astmapper.ShardAnnotation{
								Shard: 0,
								Of:    2,
							},
							SampleExpr: &vectorAggregationExpr{
								grouping:  &grouping{},
								operation: OpTypeCount,
								left: &rangeAggregationExpr{
									operation: OpTypeRate,
									left: &logRange{
										left: &matchersExpr{
											matchers: []*labels.Matcher{
												mustNewMatcher(labels.MatchEqual, "foo", "bar"),
											},
										},
										interval: 5 * time.Minute,
									},
								},
							},
						},
						next: &ConcatSampleExpr{
							SampleExpr: DownstreamSampleExpr{
								shard: &astmapper.ShardAnnotation{
									Shard: 1,
									Of:    2,
								},
								SampleExpr: &vectorAggregationExpr{
									grouping:  &grouping{},
									operation: OpTypeCount,
									left: &rangeAggregationExpr{
										operation: OpTypeRate,
										left: &logRange{
											left: &matchersExpr{
												matchers: []*labels.Matcher{
													mustNewMatcher(labels.MatchEqual, "foo", "bar"),
												},
											},
											interval: 5 * time.Minute,
										},
									},
								},
							},
							next: nil,
						},
					},
				},
			},
		},
	} {
		t.Run(tc.in, func(t *testing.T) {
			ast, err := ParseExpr(tc.in)
			require.Equal(t, tc.err, err)

			mapped, err := m.Map(ast)

			require.Equal(t, tc.err, err)
			require.Equal(t, tc.expr.String(), mapped.String())
			require.Equal(t, tc.expr, mapped)
		})
	}
}
