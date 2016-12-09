package multiparty

import (
	"errors"
	"fmt"
	"sort"
	//	"reflect"
)

// This package follows the descriptions of Multiparty Session Types as
// discussed in Honda et al. 2008 (POPL).

// GLOBAL TYPES

type GlobalType interface {
	isWellFormed() bool
	Prefixes() [][]Prefix
	Participants() []Participant
	channels() ChannelSet
	Project(p Participant) (LocalType, error)
	equals(t GlobalType) bool
}

// TODO these kinds of non-keyed types are super annoying because we have to use
// reflection to get the value out :S
type Participant string

type Prefix struct {
	P1, P2   Participant
	PChannel Channel
}

type ValueType struct {
	ValuePrefix Prefix
	Value       Sort
	ValueNext   GlobalType
}

type BranchingType struct {
	BranchPrefix Prefix
	Branches     map[string]GlobalType
}

type RecursiveType struct {
	Bind NameType
	Body GlobalType
}

func contains(p Participant, slice []Participant) bool {
	for _, element := range slice {
		if element == p {
			return true
		}
	}
	return false
}

func disjoint(a, b []Participant) bool {
	// Though this operator could be optimized, we are only dealing with the specification here.
	for _, x := range a {
		if contains(x, b) {
			return false
		}
	}
	for _, x := range b {
		if contains(x, a) {
			return false
		}
	}
	return true
}

func count(participants ParticipantSet) int {
	ans := 0
	sort.Sort(participants)
	if len(participants) < 2 {
		return len(participants)
	}
	last := participants[0]
	ans = 1
	for _, x := range participants {
		if x != last {
			last = x
			ans++
		}
	}
	return ans
}

type Channel string
type Sort string

type ParticipantSet []Participant

func (ps ParticipantSet) Len() int {
	return len(ps)
}

func (ps ParticipantSet) Less(i, j int) bool {
	return ps[i] < ps[j]
}

func (ps ParticipantSet) Swap(i, j int) {
	temp := ps[i]
	ps[i] = ps[j]
	ps[j] = temp
}

type ChannelSet []Channel

func (cs ChannelSet) Len() int {
	return len(cs)
}

func (cs ChannelSet) Less(i, j int) bool {
	return cs[i] < cs[j]
}

func (cs ChannelSet) Swap(i, j int) {
	temp := cs[i]
	cs[i] = cs[j]
	cs[j] = temp
}

func (cs ChannelSet) equals(b ChannelSet) bool {
	if cs.Len() != b.Len() {
		return false
	}
	for k := range cs {
		if cs[k] != b[k] {
			return false
		}
	}
	return true
}

func (slice ChannelSet) contains(p Channel) bool {
	for _, element := range slice {
		if element == p {
			return true
		}
	}
	return false
}

func (a ChannelSet) uniquify() ChannelSet {
	// Makes contents of the channelset actually unique.
	if a.Len() < 1 {
		return a
	}
	var ans ChannelSet
	ans = append(make([]Channel, 0, a.Len()), a[0])
	sort.Sort(ans)
	var last Channel
	last = a[0]
	for _, ch := range a {
		if ch != last {
			ans = append(ans, ch)
			last = ch
		}
	}
	return ans

}

func (a ChannelSet) disjoint(b ChannelSet) bool {
	// Though this operator could be optimized, we are only dealing with the specification here.
	for _, x := range a {
		if b.contains(x) {
			return false
		}
	}
	for _, x := range b {
		if a.contains(x) {
			return false
		}
	}
	return true
}

func (p Prefix) participants() []Participant {
	ans := make([]Participant, 2, 2)
	ans[0] = p.P1
	ans[1] = p.P2
	return ans
}

func SingletonValue(s Sort) []Sort {
	ans := make([]Sort, 1, 1)
	ans[0] = s
	return ans
}

func (t ValueType) isWellFormed() bool {
	return t.ValueNext.isWellFormed()
}

func (t ValueType) Prefixes() [][]Prefix {
	current := append(make([]Prefix, 0, 1), t.ValuePrefix)
	ans := append(make([][]Prefix, 0, 1), current)
	for _, prefix := range t.ValueNext.Prefixes() {
		ans = append(ans, append(current, prefix...))
	}
	return ans
}

func (t ValueType) Participants() []Participant {
	return append(t.ValuePrefix.participants(), t.ValueNext.Participants()...)
}

func (t ValueType) Project(p Participant) (LocalType, error) {
	ans, err := t.ValueNext.Project(p)
	if err != nil {
		return nil, err
	} else if t.ValuePrefix.P1 == p {
		ans = LocalSendType{Channel: t.ValuePrefix.PChannel, Value: t.Value, Next: ans}
	} else if t.ValuePrefix.P2 == p {
		ans = LocalReceiveType{Channel: t.ValuePrefix.PChannel, Value: t.Value, Next: ans}
	}
	return ans, err
}

func (t ValueType) equals(g GlobalType) bool {
	switch g.(type) {
	case ValueType:
		gt := g.(ValueType)
		return gt.ValuePrefix == t.ValuePrefix && (gt.Value == t.Value) && t.ValueNext.equals(gt.ValueNext)
	}
	return false
}

func (t ValueType) channels() ChannelSet {
	return append(t.ValueNext.channels(), t.ValuePrefix.PChannel)
}

func (t BranchingType) channels() ChannelSet {
	ans := make([]Channel, 0, 0)
	for _, v := range t.Branches {
		ans = append(ans, v.channels()...)
	}
	return append(ans, t.BranchPrefix.PChannel)
}

func (t BranchingType) Prefixes() [][]Prefix {
	base := append(make([]Prefix, 0, 1), t.BranchPrefix)
	ans := make([][]Prefix, 0, 1)
	for _, branch := range t.Branches {
		branches := branch.Prefixes()
		for _, branch := range branches {
			ans = append(ans, append(base, branch...))
		}
	}
	return ans
}

func (t BranchingType) isWellFormed() bool {
	for _, value := range t.Branches {
		if !value.isWellFormed() {
			return false
		}
	}
	return true
}

func (b BranchingType) Participants() []Participant {
	ans := b.BranchPrefix.participants()
	for _, val := range b.Branches {
		ans = append(ans, val.Participants()...)
	}
	return ans
}

func (b BranchingType) Project(p Participant) (LocalType, error) {
	branches := make(map[string]LocalType)

	for key, value := range b.Branches {
		candidate, err := value.Project(p)
		if err != nil {
			return nil, err
		}
		branches[key] = candidate
	}

	unique := func(branches map[string]LocalType) bool {
		var prevValue LocalType

		// preset with first value in the range

		for _, prevValue = range branches {
			break
		}

		for _, value := range branches {
			if !prevValue.Equals(value) {
				return false
			}
		}
		return true
	}

	if b.BranchPrefix.P1 == p {
		return LocalSelectionType{Channel: b.BranchPrefix.PChannel, Branches: branches}, nil
	} else if b.BranchPrefix.P2 == p {
		return LocalBranchingType{Channel: b.BranchPrefix.PChannel, Branches: branches}, nil
	} else if unique(branches) {
		for _, branch := range branches {
			return branch, nil
		}
	}
	return nil, errors.New("projection undefined")
}

func (t BranchingType) equals(g GlobalType) bool {
	switch g.(type) {
	case BranchingType:
		gt := g.(BranchingType)
		for x := range t.Branches {
			if !t.Branches[x].equals(gt.Branches[x]) {
				return false
			}
		}
		return t.BranchPrefix == gt.BranchPrefix
	}
	return false
}

type ParallelType struct {
	a, b GlobalType
}

func MakeParallelType(a GlobalType, b GlobalType) ParallelType {
	return ParallelType{a, b}
}

func (t ParallelType) channels() ChannelSet {
	return append(t.a.channels(), t.b.channels()...)
}

func (t ParallelType) Prefixes() [][]Prefix {
	return append(t.a.Prefixes(), t.b.Prefixes()...)
}

func (t ParallelType) isWellFormed() bool {
	return t.a.isWellFormed() && t.b.isWellFormed()
}

func (t ParallelType) Participants() []Participant {
	return append(t.a.Participants(), t.b.Participants()...)
}

func (t ParallelType) Project(p Participant) (LocalType, error) {
	if contains(p, t.a.Participants()) {
		if contains(p, t.b.Participants()) {
			return nil, errors.New("projection undefined")
		}
		return t.a.Project(p)
	}
	if contains(p, t.b.Participants()) {
		return t.b.Project(p)
	}
	ans := LocalEndType{}
	return ans, nil
}

func (t ParallelType) equals(g GlobalType) bool {
	switch g.(type) {
	case ParallelType:
		gt := g.(ParallelType)
		return t.a.equals(gt.a) && t.b.equals(gt.b)
	}
	return false
}

type NameType string

func (t NameType) isWellFormed() bool {
	return true
}

func (t NameType) Prefixes() [][]Prefix {
	return make([][]Prefix, 0, 0)
}

func (t NameType) Participants() []Participant {
	return make([]Participant, 0, 0)
}

func (t NameType) Project(p Participant) (LocalType, error) {
	return LocalNameType(t), nil
}

func (t NameType) channels() ChannelSet {
	return make([]Channel, 0, 0)
}

func (t NameType) equals(g GlobalType) bool {
	switch g.(type) {
	case NameType:
		return t == g.(NameType)
	}
	return false
}

func (t RecursiveType) channels() ChannelSet {
	return t.Body.channels()
}

func (t RecursiveType) isWellFormed() bool {
	return t.Body.isWellFormed()
}

func (t RecursiveType) Prefixes() [][]Prefix {
	return t.Body.Prefixes()
}

func (t RecursiveType) Participants() []Participant {
	return t.Body.Participants()
}

func (t RecursiveType) Project(p Participant) (LocalType, error) {
	body, err := t.Body.Project(p)
	if err != nil {
		return nil, err
	}
	return LocalRecursiveType{Bind: LocalNameType(t.Bind), Body: body}, nil
}

func (t RecursiveType) equals(g GlobalType) bool {
	switch g.(type) {
	case RecursiveType:
		gt := g.(RecursiveType)
		return t.Bind.equals(gt.Bind) && t.Body.equals(gt.Body)
	}
	return false
}

type EndType struct{}

func (t EndType) isWellFormed() bool {
	return true
}

func (t EndType) channels() ChannelSet {
	return make([]Channel, 0, 0)
}
func (t EndType) Prefixes() [][]Prefix {
	return make([][]Prefix, 0, 0)
}

func (t EndType) Participants() []Participant {
	return make([]Participant, 0, 0)
}

func (t EndType) Project(p Participant) (LocalType, error) {
	return LocalEndType(t), nil
}

func (t EndType) equals(g GlobalType) bool {
	switch g.(type) {
	case EndType:
		return true
	}
	return false
}

func linear(original_gt GlobalType) bool {
	gt := unfold(original_gt, make(map[NameType]GlobalType))
	return linearInternal(gt, make([]Prefix, 0, 0))
}

//Definition 4.2
func coherent(original_gt GlobalType) bool {
	if !linear(original_gt) {
		return false
	}
	gt := unfold(original_gt, make(map[NameType]GlobalType))
	for _, p := range gt.Participants() {
		_, err := gt.Project(p)
		if err != nil {
			return false
		}
	}
	return true
}

func (n1 Prefix) II(n2 Prefix) bool {
	if n1.P2 != n2.P2 {
		fmt.Printf("DEP: II doesn't hold. expects equal P2 among %+v and %+v\n", n1, n2)
		return false
	}
	if n1.PChannel != n2.PChannel && n1.P1 != n2.P1 {
		//second if condition given on the tech report.
		return true
	}
	fmt.Printf("DEP: II doesn't hold. expects different channels among %+v and %+v\n", n1, n2)

	return false
}

func (n1 Prefix) IO(n2 Prefix) bool {
	if n1.P2 != n2.P1 {
		fmt.Printf("DEP: IO doesn't hold. expects shared participant among %+v and %+v.\n", n1, n2)
		return false
	}
	if n1.PChannel == n2.PChannel {
		fmt.Printf("DEP: IO doesn't hold. expects different channels on n1P1 and n2P2 for %+v and %+v\n", n1, n2)
		return false
	}
	return true
}

func (n1 Prefix) OO(n2 Prefix) bool {
	if n1.P1 != n2.P1 {
		fmt.Printf("DEP: OO doesn't hold. expects same P1 among %+v and %+v.\n", n1, n2)
		return false
	}
	if n1.PChannel != n2.PChannel && n1.P2 != n2.P2 {
		//extra if confition derived from tech report.
		//OO holds subject to p1 \neq p2 => k1 \neq k2,
		// for pfx(n1) = p -> p1: k1 and pfx(n2) = p -> p2 : k2
		// (Tech report at http://www.doc.ic.ac.uk/~pmalo/research/papers/buffer-communication-analysis.pdf)
		fmt.Printf("DEP: OO doesn't hold. expects shared channel among %v and %v.\n", n1, n2)
		return false
	}
	return true
}

func filter_shared_channel(lessthan []Prefix, filter Prefix) []Prefix {
	ans := make([]Prefix, 0, 0)
	for _, prefix := range lessthan {
		if prefix.PChannel == filter.PChannel { //&& prefix != filter{
			ans = append(ans, prefix)
		}
	}
	return ans
}

func linearInternal(gt GlobalType, lessthan []Prefix) bool {
	/*
		overall implementation idea:
		since we already have unwrapped, we only need to locally check and there's a finite amount of nodes to explore (we already removed the cycles via unfold)

		-
	*/
	switch gt.(type) {
	case ValueType:
		t := gt.(ValueType)
		filtered_lessthan := filter_shared_channel(lessthan, t.ValuePrefix)
		if !(InputDependency(filtered_lessthan, t.ValuePrefix) && OutputDependency(filtered_lessthan, t.ValuePrefix)) {
			fmt.Printf("ERROR: Failed to collect dependencies for ValueType %+v\n", t)
			return false
		}
		linearInternal(t.ValueNext, append(lessthan, t.ValuePrefix))
	case BranchingType:
		t := gt.(BranchingType)
		filtered_lessthan := filter_shared_channel(lessthan, t.BranchPrefix)
		if !(InputDependency(filtered_lessthan, t.BranchPrefix) &&
			OutputDependency(filtered_lessthan, t.BranchPrefix)) {
			fmt.Printf("ERROR: Failed to collect dependencies for BranchingType %+v\n", t)
			return false
		}
		new_lessthan := append(lessthan, t.BranchPrefix)
		for _, value := range t.Branches {
			if !linearInternal(value, new_lessthan) {
				return false
			}
		}
	case ParallelType:
		t := gt.(ParallelType)
		for _, prefixes := range t.b.Prefixes() {
			//fmt.Printf("DEBUG: Checking Parallel linearity with %+v and %+v\n",lessthan, prefixes)
			if !linearInternal(t.a, append(lessthan, prefixes...)) {
				return false
			}
		}
		for _, prefixes := range t.a.Prefixes() {
			if !linearInternal(t.b, append(lessthan, prefixes...)) {
				return false
			}
		}
	case RecursiveType:
		t := gt.(RecursiveType)
		//fmt.Printf("DEBUG: Entering body of type %+v\n", t)
		return linearInternal(t.Body, lessthan)
	case NameType:
	case EndType:
	}
	return true
}

func InputDependency(firsts []Prefix, last Prefix) bool {
	if len(firsts) < 1 {
		return true
	}
	for i := 0; i < len(firsts)-1; i++ {
		if !(firsts[i].II(last) || firsts[i].IO(last)) {
			fmt.Printf("ERROR: Broke input dependency with %v and %v\n", firsts[i], last)
			return false
		}
	}
	if !firsts[len(firsts)-1].II(last) {
		fmt.Printf("Error: Broke input dependency last-II invariant between %v and %v.\n", firsts[len(firsts)-1], last)
		return false
	}
	return true
}

func OutputDependency(firsts []Prefix, last Prefix) bool {
	if len(firsts) < 1 {
		return true
	}
	for _, first := range firsts {
		if !(first.IO(last) || first.OO(last)) {
			fmt.Printf("Error: Broke output dependency invariant between %v and %v.\n", first, last)
			return false
		}
	}
	return true
}

func unfold(gt GlobalType, env map[NameType]GlobalType) GlobalType {
	switch gt.(type) {
	case ValueType:
		t := gt.(ValueType)
		return ValueType{ValuePrefix: t.ValuePrefix, Value: t.Value, ValueNext: unfold(t.ValueNext, env)}
	case BranchingType:
		t := gt.(BranchingType)
		branches := make(map[string]GlobalType)
		for key, value := range t.Branches {
			branches[key] = unfold(value, env)
		}
		return BranchingType{BranchPrefix: t.BranchPrefix, Branches: branches}
	case ParallelType:
		t := gt.(ParallelType)
		return ParallelType{a: unfold(t.a, env), b: unfold(t.b, env)}
	case RecursiveType:
		t := gt.(RecursiveType)
		if val, ok := env[t.Bind]; ok {
			if val != t {
				//name hiding!
				old_val := val
				env[t.Bind] = t
				ans := RecursiveType{Bind: t.Bind, Body: unfold(t.Body, env)}
				env[t.Bind] = old_val
				return ans
			} else {
				//I already unfolded once. then return (avoid infinite loop)
				return t
			}
		}
		env[t.Bind] = t
		return RecursiveType{Bind: t.Bind, Body: unfold(t.Body, env)}
	case NameType:
		t := gt.(NameType)
		if val, ok := env[t]; ok {
			return val
		}
		return t
	case EndType:
		return gt
	}
	return nil
}

// LOCAL TYPES

type LocalType interface {
	Equals(t LocalType) bool
	Substitute(u LocalNameType, t LocalType) LocalType
}

type ProjectionType struct {
	// Originall, Type @ participant
	T           LocalType
	participant Participant
}

func (t ProjectionType) Substitute(u LocalNameType, tsub LocalType) LocalType {
	ret := t
	ret.T = t.T.Substitute(u, tsub)
	return ret
}

func (t ProjectionType) Equals(l LocalType) bool {
	switch l.(type) {
	case ProjectionType:
		lt := l.(ProjectionType)
		return t.participant == lt.participant && t.T.Equals(lt.T)
	}
	return false
}

func findProjection(participant Participant, projections []ProjectionType) (ProjectionType, error) {
	var ans ProjectionType
	for _, proj := range projections {
		if participant == proj.participant {
			return proj, nil
		}
	}
	return ans, errors.New(fmt.Sprintf("Could not find projection for participant %+v in set %+v", participant, projections))
}

type LocalSendType struct {
	Channel Channel
	Value   Sort
	Next    LocalType
}

func (t LocalSendType) Substitute(u LocalNameType, tsub LocalType) LocalType {
	ret := t
	ret.Next = t.Next.Substitute(u, tsub)
	return ret
}

func (t LocalSendType) Equals(l LocalType) bool {
	switch l.(type) {
	case LocalSendType:
		lt := l.(LocalSendType)
		return t.Channel == lt.Channel && (t.Value == lt.Value) && t.Next.Equals(lt.Next)
	}
	return false
}

type LocalReceiveType struct {
	Channel Channel
	Value   Sort
	Next    LocalType
}

func (t LocalReceiveType) Substitute(u LocalNameType, tsub LocalType) LocalType {
	ret := t
	ret.Next = t.Next.Substitute(u, tsub)
	return ret
}

func (t LocalReceiveType) Equals(l LocalType) bool {
	switch l.(type) {
	case LocalReceiveType:
		lt := l.(LocalReceiveType)
		return t.Channel == lt.Channel && (t.Value == lt.Value) && t.Next.Equals(lt.Next)
	}
	return false
}

type LocalSelectionType struct {
	// k \oplus
	Channel  Channel
	Branches map[string]LocalType
}

func (t LocalSelectionType) Substitute(u LocalNameType, tsub LocalType) LocalType {
	ret := t
	for k, branchType := range t.Branches {
		ret.Branches[k] = branchType.Substitute(u, tsub)
	}
	return ret
}

func (t LocalSelectionType) Equals(l LocalType) bool {
	switch l.(type) {
	case LocalSelectionType:
		lt := l.(LocalSelectionType)
		for k := range t.Branches {
			if !t.Branches[k].Equals(lt.Branches[k]) {
				return false
			}
		}
		return t.Channel == lt.Channel
	}
	return false
}

type LocalBranchingType struct {
	// k &
	Channel  Channel
	Branches map[string]LocalType
}

func (t LocalBranchingType) Substitute(u LocalNameType, tsub LocalType) LocalType {
	ret := t
	for k, branchType := range t.Branches {
		ret.Branches[k] = branchType.Substitute(u, tsub)
	}
	return ret
}

func (t LocalBranchingType) Equals(l LocalType) bool {
	switch l.(type) {
	case LocalBranchingType:
		lt := l.(LocalBranchingType)
		for k := range t.Branches {
			if !t.Branches[k].Equals(lt.Branches[k]) {
				return false
			}
		}
		return t.Channel == lt.Channel
	}
	return false
}

type LocalNameType string

func (t LocalNameType) GoString() string {
	return fmt.Sprintf("multiparty.LocalNameType(\"%s\")", t)
}

func (t LocalNameType) Substitute(u LocalNameType, tsub LocalType) LocalType {
	if u == t {
		return tsub
	} else {
		return u
	}
}

func (t LocalNameType) Equals(l LocalType) bool {
	switch l.(type) {
	case LocalNameType:
		return t == l.(LocalNameType)
	}
	return false
}

type LocalRecursiveType struct {
	Bind LocalNameType
	Body LocalType
}

func (t LocalRecursiveType) UnfoldOneLevel() LocalType {
	return t.Body.Substitute(t.Bind, t)
}

func (t LocalRecursiveType) Substitute(u LocalNameType, tsub LocalType) LocalType {
	//Don't substitute if we're shadowing
	if u == t.Bind {
		return u
	} else {
		ret := t
		ret.Body = t.Body.Substitute(u, tsub)
		return ret
	}
}

func (t LocalRecursiveType) Equals(l LocalType) bool {
	switch l.(type) {
	case LocalRecursiveType:
		lt := l.(LocalRecursiveType)
		return t.Bind.Equals(lt.Bind) && t.Body.Equals(lt.Body)
	}
	return false
}

type LocalEndType struct{}

func (t LocalEndType) Substitute(u LocalNameType, tsub LocalType) LocalType {
	return t
}

func (t LocalEndType) Equals(l LocalType) bool {
	switch l.(type) {
	case LocalEndType:
		return true
	}
	return false
}

// end of local types and definition of projection.
