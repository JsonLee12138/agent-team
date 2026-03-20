package governance

import "time"

type GateInput struct {
	TaskPacket      TaskPacket
	Index           Index
	LoadedRules     RuleLoadResult
	ArchivedTicket  *ArchivedExceptionTicket
	ConsumeArchived func(ticket *ArchivedExceptionTicket, taskID, owner string) error
}

func EvaluateGate(input GateInput) GateResult {
	packet := input.TaskPacket

	if !HasIndexEntry(input.Index, packet.TaskID) {
		return DeclaredReferenceNotFound(packet.TaskID, packet.ModuleID, packet.TaskID)
	}

	missing := MissingReferences(input.Index, packet.DeclaredReferences)
	if len(missing) > 0 {
		return DeclaredReferenceNotFound(packet.TaskID, packet.ModuleID, missing[0])
	}

	if len(input.LoadedRules.Conflicts) > 0 {
		return RuleOverrideConflict(packet.TaskID, packet.ModuleID, input.LoadedRules.Conflicts[0])
	}

	if packet.UsesArchivedInput {
		if input.ArchivedTicket == nil {
			return ArchivedBlocked(packet.TaskID, packet.ModuleID)
		}
		consume := input.ConsumeArchived
		if consume == nil {
			consume = func(ticket *ArchivedExceptionTicket, taskID, owner string) error {
				return ConsumeArchivedException(ticket, taskID, owner, TimeNowUTC())
			}
		}
		if err := consume(input.ArchivedTicket, packet.TaskID, packet.Owner); err != nil {
			return ArchivedBlocked(packet.TaskID, packet.ModuleID)
		}
	}

	return PassGateResult()
}

var TimeNowUTC = func() time.Time {
	return time.Now().UTC()
}
