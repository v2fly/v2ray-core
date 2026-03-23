package rrpitBidirectionalSession

import "testing"

type bidirectionalFuzzConfig struct {
	maxSourceLossPerLane    int
	maxRepairLossPerLane    int
	maxControlLossPerSender int
	maxLatencyTicks         int
	maxMessagesPerSide      int
	channelHistoryCapacity  int
}

type bidirectionalFuzzSeed struct {
	channelCountRaw  uint8
	shardSizeRaw     uint8
	maxDataShardsRaw uint8
	sourceLossRaw    uint8
	repairLossRaw    uint8
	controlLossRaw   uint8
	latencyRaw       uint8
	aOddChannelIDs   bool
	bOddChannelIDs   bool
	schedule         []byte
	aPayload         []byte
	bPayload         []byte
}

func FuzzBidirectionalSessionRoundTrip(f *testing.F) {
	fuzzBidirectionalSessionRoundTrip(f, bidirectionalFuzzConfig{
		maxSourceLossPerLane:    3,
		maxRepairLossPerLane:    3,
		maxControlLossPerSender: 3,
		maxLatencyTicks:         4,
		maxMessagesPerSide:      6,
		channelHistoryCapacity:  32,
	}, []bidirectionalFuzzSeed{
		{
			channelCountRaw:  1,
			shardSizeRaw:     16,
			maxDataShardsRaw: 1,
			sourceLossRaw:    0,
			repairLossRaw:    0,
			controlLossRaw:   0,
			latencyRaw:       0,
			aOddChannelIDs:   true,
			bOddChannelIDs:   false,
			schedule:         []byte{0, 1, 2, 3},
			aPayload:         []byte("alpha"),
			bPayload:         []byte("bravo"),
		},
		{
			channelCountRaw:  2,
			shardSizeRaw:     24,
			maxDataShardsRaw: 2,
			sourceLossRaw:    1,
			repairLossRaw:    1,
			controlLossRaw:   1,
			latencyRaw:       2,
			aOddChannelIDs:   false,
			bOddChannelIDs:   true,
			schedule:         []byte{5, 4, 3, 2, 1},
			aPayload:         []byte("hello-world"),
			bPayload:         []byte("goodbye-moon"),
		},
		{
			channelCountRaw:  3,
			shardSizeRaw:     32,
			maxDataShardsRaw: 3,
			sourceLossRaw:    2,
			repairLossRaw:    1,
			controlLossRaw:   2,
			latencyRaw:       3,
			aOddChannelIDs:   true,
			bOddChannelIDs:   false,
			schedule:         []byte{9, 7, 5, 3, 1, 0},
			aPayload:         []byte("abcdefghijk"),
			bPayload:         []byte("lmnopqrstuv"),
		},
	})
}

func FuzzBidirectionalSessionRoundTripHighLoss(f *testing.F) {
	fuzzBidirectionalSessionRoundTrip(f, bidirectionalFuzzConfig{
		maxSourceLossPerLane:    7,
		maxRepairLossPerLane:    6,
		maxControlLossPerSender: 6,
		maxLatencyTicks:         8,
		maxMessagesPerSide:      12,
		channelHistoryCapacity:  128,
	}, []bidirectionalFuzzSeed{
		{
			channelCountRaw:  1,
			shardSizeRaw:     32,
			maxDataShardsRaw: 3,
			sourceLossRaw:    6,
			repairLossRaw:    5,
			controlLossRaw:   4,
			latencyRaw:       7,
			aOddChannelIDs:   true,
			bOddChannelIDs:   false,
			schedule:         []byte{8, 6, 7, 5, 3, 0, 9},
			aPayload:         []byte("high-loss-alpha-seed"),
			bPayload:         []byte("high-loss-bravo-seed"),
		},
		{
			channelCountRaw:  2,
			shardSizeRaw:     30,
			maxDataShardsRaw: 2,
			sourceLossRaw:    7,
			repairLossRaw:    6,
			controlLossRaw:   5,
			latencyRaw:       8,
			aOddChannelIDs:   false,
			bOddChannelIDs:   true,
			schedule:         []byte{1, 4, 1, 4, 2, 1, 3, 5, 8},
			aPayload:         []byte("the-quick-brown-fox-jumps"),
			bPayload:         []byte("over-the-lazy-dog-twice"),
		},
		{
			channelCountRaw:  3,
			shardSizeRaw:     28,
			maxDataShardsRaw: 3,
			sourceLossRaw:    5,
			repairLossRaw:    6,
			controlLossRaw:   6,
			latencyRaw:       6,
			aOddChannelIDs:   true,
			bOddChannelIDs:   false,
			schedule:         []byte{2, 7, 1, 8, 2, 8, 1, 8},
			aPayload:         []byte("packets-may-drop-many-times-before-repair"),
			bPayload:         []byte("control-latency-and-loss-should-still-converge"),
		},
	})
}

func fuzzBidirectionalSessionRoundTrip(
	f *testing.F,
	config bidirectionalFuzzConfig,
	seeds []bidirectionalFuzzSeed,
) {
	for _, seed := range seeds {
		f.Add(
			seed.channelCountRaw,
			seed.shardSizeRaw,
			seed.maxDataShardsRaw,
			seed.sourceLossRaw,
			seed.repairLossRaw,
			seed.controlLossRaw,
			seed.latencyRaw,
			seed.aOddChannelIDs,
			seed.bOddChannelIDs,
			seed.schedule,
			seed.aPayload,
			seed.bPayload,
		)
	}

	f.Fuzz(func(
		t *testing.T,
		channelCountRaw uint8,
		shardSizeRaw uint8,
		maxDataShardsRaw uint8,
		sourceLossRaw uint8,
		repairLossRaw uint8,
		controlLossRaw uint8,
		latencyRaw uint8,
		aOddChannelIDs bool,
		bOddChannelIDs bool,
		schedule []byte,
		aPayload []byte,
		bPayload []byte,
	) {
		channelCount := int(channelCountRaw%3) + 1
		maxDataShardsPerLane := int(maxDataShardsRaw%3) + 1
		shardSize := int(shardSizeRaw%25) + 8
		profile := bidirectionalTickedNetworkProfile{
			maxSourceLossPerLane:    boundedFuzzValue(sourceLossRaw, config.maxSourceLossPerLane),
			maxRepairLossPerLane:    boundedFuzzValue(repairLossRaw, config.maxRepairLossPerLane),
			maxControlLossPerSender: boundedFuzzValue(controlLossRaw, config.maxControlLossPerSender),
			maxLatencyTicks:         boundedFuzzValue(latencyRaw, config.maxLatencyTicks),
		}
		maxMessageLen := shardSize - 2
		if maxMessageLen < 1 {
			t.Skip()
		}

		aMessages := fuzzMessagesFromBytes(aPayload, maxMessageLen, config.maxMessagesPerSide, 'a')
		bMessages := fuzzMessagesFromBytes(bPayload, maxMessageLen, config.maxMessagesPerSide, 'b')

		maxBufferedLanes := max(
			2,
			max(
				expectedLaneCountForConfig(len(aMessages), maxDataShardsPerLane),
				expectedLaneCountForConfig(len(bMessages), maxDataShardsPerLane),
			)+1,
		)

		h := newBidirectionalIntegrationHarness(t, bidirectionalIntegrationConfig{
			name:                   "fuzz",
			channelCount:           channelCount,
			shardSize:              shardSize,
			maxDataShardsPerLane:   maxDataShardsPerLane,
			maxBufferedLanes:       maxBufferedLanes,
			channelHistoryCapacity: config.channelHistoryCapacity,
			aOddChannelIDs:         aOddChannelIDs,
			bOddChannelIDs:         bOddChannelIDs,
		})

		h.bootstrapAndLearn()
		h.sendMessages(h.a, aMessages...)
		h.sendMessages(h.b, bMessages...)
		settleTicks := h.settleTicksForProfile(profile)
		h.runUntilExactDelivery(
			bMessages,
			aMessages,
			newBidirectionalTickedNetwork(profile, schedule),
			h.maxTicksForProfile(len(aMessages), len(bMessages), profile, settleTicks),
			settleTicks,
		)

		h.assertReceived(h.a, bMessages)
		h.assertReceived(h.b, aMessages)
		h.assertControlCallbacksObserved()
	})
}

func boundedFuzzValue(raw uint8, maxValue int) int {
	if maxValue <= 0 {
		return 0
	}
	return int(raw % uint8(maxValue+1))
}

func fuzzMessagesFromBytes(data []byte, maxMessageLen int, maxMessages int, fallback byte) [][]byte {
	if maxMessageLen < 1 || maxMessages < 1 {
		return [][]byte{{fallback}}
	}
	if len(data) == 0 {
		return [][]byte{{fallback}}
	}

	messages := make([][]byte, 0, maxMessages)
	cursor := 0
	for cursor < len(data) && len(messages) < maxMessages {
		size := int(data[cursor]%uint8(maxMessageLen)) + 1
		cursor += 1

		remaining := len(data) - cursor
		if remaining <= 0 {
			messages = append(messages, []byte{fallback})
			continue
		}
		if size > remaining {
			size = remaining
		}
		messages = append(messages, append([]byte(nil), data[cursor:cursor+size]...))
		cursor += size
	}
	if len(messages) == 0 {
		return [][]byte{{fallback}}
	}
	return messages
}

func rotateBidirectionalFrames(frames []bidirectionalIntegrationFrame, offset int) []bidirectionalIntegrationFrame {
	if len(frames) == 0 || offset == 0 {
		return frames
	}
	offset %= len(frames)
	if offset < 0 {
		offset += len(frames)
	}

	rotated := append([]bidirectionalIntegrationFrame{}, frames[offset:]...)
	rotated = append(rotated, frames[:offset]...)
	return rotated
}
