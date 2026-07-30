package parser

import (
	"reflect"

	"github.com/sqlc-dev/teesql/ast"
)

// frag returns the Fragment of an AST node, or nil for nil nodes and values
// that are not AST nodes.
func frag(n any) *ast.Fragment {
	s, ok := n.(interface{ Frag() *ast.Fragment })
	if !ok {
		return nil
	}
	if v := reflect.ValueOf(n); v.Kind() == reflect.Pointer && v.IsNil() {
		return nil
	}
	return s.Frag()
}

// neverDerive lists node types that ScriptDom sometimes emits without
// position information (typically synthesized defaults for omitted optional
// syntax). For these, a span is only emitted when the parser recorded one
// explicitly; it is never derived from children.
var neverDerive = map[string]bool{
	"AlterServerConfigurationExternalAuthenticationOption": true,
	"AlterTableDropTableElement":                           true,
	"AuthenticationPayloadOption":                          true,
	"BrokerPriorityParameter":                              true,
	"CharacterSetPayloadOption":                            true,
	"DurabilityTableOption":                                true,
	"EnabledDisabledPayloadOption":                         true,
	"EncryptionPayloadOption":                              true,
	"EventRetentionSessionOption":                          true,
	"FetchType":                                            true,
	"FileGroupDefinition":                                  true,
	"FullTextStopListAction":                               true,
	"GenericConfigurationOption":                           true,
	"GraphRecursiveMatchQuantifier":                        true,
	"GridParameter":                                        true,
	"Identifier":                                           true,
	"IdentifierLiteral":                                    true,
	"IndexType":                                            true,
	"LedgerViewOption":                                     true,
	"ListenerIPEndpointProtocolOption":                     true,
	"LiteralEndpointProtocolOption":                        true,
	"LoginTypePayloadOption":                               true,
	"MaxDispatchLatencySessionOption":                      true,
	"MemoryPartitionSessionOption":                         true,
	"RemoteDataArchiveDbCredentialSetting":                 true,
	"RemoteDataArchiveDbFederatedServiceAccountSetting":    true,
	"RemoteDataArchiveDbServerSetting":                     true,
	"RetentionPeriodDefinition":                            true,
	"RolePayloadOption":                                    true,
	"SchemaPayloadOption":                                  true,
	"SessionTimeoutPayloadOption":                          true,
	"StatementList":                                        true,
	"StringLiteral":                                        true,
	"TableClusteredIndexType":                              true,
	"TableIndexOption":                                     true,
	"TemporalClause":                                       true,
	"UserLoginOption":                                      true,
	"WsdlPayloadOption":                                    true,
}

// addSpan copies a node's recorded source span into its JSON representation.
// When a node carries no recorded span (wrapper nodes synthesized during
// marshaling, such as MultiPartIdentifier), the span is derived from the
// node's already-marshaled children, unless the type is known to appear
// without positions in ScriptDom output.
func addSpan(n jsonNode, f *ast.Fragment) jsonNode {
	if n == nil {
		return n
	}
	if f != nil && f.HasSpan() {
		n["StartOffset"] = f.StartOffset
		n["FragmentLength"] = f.FragmentLength
		n["StartLine"] = f.StartLine
		n["StartColumn"] = f.StartColumn
		return n
	}
	t, _ := n["$type"].(string)
	if t == "" || neverDerive[t] {
		return n
	}
	if _, done := n["StartOffset"]; done {
		return n
	}
	minS, maxE, line, col := -1, -1, 0, 0
	var consider func(v any)
	consider = func(v any) {
		switch c := v.(type) {
		case jsonNode:
			so, ok := c["StartOffset"].(int)
			if !ok {
				return
			}
			fl, _ := c["FragmentLength"].(int)
			if minS < 0 || so < minS {
				minS = so
				line, _ = c["StartLine"].(int)
				col, _ = c["StartColumn"].(int)
			}
			if so+fl > maxE {
				maxE = so + fl
			}
		case map[string]any:
			consider(jsonNode(c))
		case []jsonNode:
			for _, e := range c {
				consider(e)
			}
		case []any:
			for _, e := range c {
				consider(e)
			}
		}
	}
	for k, v := range n {
		if k == "$type" {
			continue
		}
		consider(v)
	}
	if minS >= 0 {
		n["StartOffset"] = minS
		n["FragmentLength"] = maxE - minS
		n["StartLine"] = line
		n["StartColumn"] = col
	}
	return n
}
