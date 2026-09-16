package main

import (
	"fmt"
)

type OpType struct {
	AchieveText bool
	Received    bool
	Comment     bool
	Gift        int
}
var opType = []OpType{
	{AchieveText: false, Received: false, Comment: false, Gift: 0},  // SW2026-0
	{AchieveText: true, Received: true, Comment: false, Gift: 0},  // SW2026-1
	{AchieveText: false, Received: false, Comment: false, Gift: 0},  // SW2026-2
	{AchieveText: false, Received: false, Comment: false, Gift: 0},  // SW2026-3
	{AchieveText: true, Received: true, Comment: true, Gift: 0},  // SW2026-4
	{AchieveText: true, Received: true, Comment: false, Gift: 10},  // SW2026-5
	{AchieveText: true, Received: true, Comment: false, Gift: 0},  // SW2026-6
	{AchieveText: false, Received: false, Comment: false, Gift: 0},  // SW2026-7
	{AchieveText: false, Received: false, Comment: false, Gift: 0},  // SW2026-8
	{AchieveText: false, Received: false, Comment: false, Gift: 0},  // SW2026-9
	{AchieveText: false, Received: false, Comment: false, Gift: 0},  // SW2026-10
	{AchieveText: false, Received: false, Comment: false, Gift: 0},  // SW2026-11
	{AchieveText: false, Received: false, Comment: false, Gift: 0},  // SW2026-12
	{AchieveText: false, Received: false, Comment: false, Gift: 0},  // SW2026-13
	{AchieveText: false, Received: false, Comment: false, Gift: 0},  // SW2026-14
	{AchieveText: false, Received: false, Comment: false, Gift: 0},  // SW2026-15
	{AchieveText: false, Received: false, Comment: false, Gift: 0},  // SW2026-16
	{AchieveText: false, Received: false, Comment: false, Gift: 0},  // SW2026-17
	{AchieveText: false, Received: false, Comment: false, Gift: 0},  // SW2026-18
	{AchieveText: false, Received: false, Comment: false, Gift: 0},  // SW2026-19
	{AchieveText: false, Received: false, Comment: false, Gift: 0},  // SW2026-20
	{AchieveText: false, Received: false, Comment: false, Gift: 0},  // SW2026-21
	{AchieveText: false, Received: false, Comment: false, Gift: 0},  // SW2026-22
	{AchieveText: false, Received: false, Comment: false, Gift: 0},  // SW2026-23
	{AchieveText: false, Received: false, Comment: false, Gift: 0},  // SW2026-24
	{AchieveText: false, Received: false, Comment: false, Gift: 0},  // SW2026-25
	{AchieveText: false, Received: false, Comment: false, Gift: 0},  // SW2026-26
	{AchieveText: false, Received: false, Comment: false, Gift: 0},  // SW2026-27
	{AchieveText: false, Received: false, Comment: false, Gift: 0},  // SW2026-28
	{AchieveText: false, Received: false, Comment: false, Gift: 0},  // SW2026-29
}

func mkSelector(mtype string, no int) (selector string, op OpType, err error) {
	switch mtype {
	case "SW2026":
		/*
		   .missions > div:nth-child(4) > ol:nth-child(2) > li:nth-child(1)
		   .missions > div:nth-child(4) > ol:nth-child(2) > li:nth-child(16)

		   .missions > div:nth-child(5) > ol:nth-child(2) > li:nth-child(1)
		   .missions > div:nth-child(5) > ol:nth-child(2) > li:nth-child(7)

		   .missions > div:nth-child(5) > ol:nth-child(4) > li:nth-child(1)
		   .missions > div:nth-child(5) > ol:nth-child(4) > li:nth-child(7)
		*/
		if no < 16 {
			selector = fmt.Sprintf(".missions > div:nth-child(4) > ol:nth-child(2) > li:nth-child(%d)", no+1)
		} else if no < 23 {
			selector = fmt.Sprintf(".missions > div:nth-child(5) > ol:nth-child(2) > li:nth-child(%d)", no-15)
		} else if no < 30 {
			selector = fmt.Sprintf(".missions > div:nth-child(5) > ol:nth-child(4) > li:nth-child(%d)", no-22)
		} else {
			err = fmt.Errorf("invalid mission number: %d", no)
			return
		}
		op = opType[no]
	default:
	}
	return
}
