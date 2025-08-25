# Glossary

## A

### **Agent**

An independent, autonomous unit in a swarm that maintains its own state (phase, frequency, energy) and coordinates with neighbors. Like a single firefly in a group, each agent follows simple rules that lead to complex collective behavior.

### **Adjustment**

The change an agent makes to its phase or frequency based on neighbor observations. Small, continuous adjustments lead to global synchronization.

### **Algorithm**

The emerge algorithm is based on the Kuramoto model, describing how coupled oscillators synchronize through local interactions.

### **Anti-Phase Synchronization**

When agents deliberately maintain different phases to avoid acting simultaneously. Used for load distribution to prevent spikes.

### **Atomic Operations**

Lock-free operations that enable agents to update their state without mutual exclusion, allowing true parallel execution.

### **Autonomy**

The property of agents making their own decisions without central control or external commands.

## B

### **Batching**

Grouping multiple operations to execute together. Emerge enables dynamic batching by synchronizing when agents act.

### **Byzantine Fault Tolerance**

The ability to handle malicious or incorrectly behaving agents (future feature).

## C

### **Client API**

The high-level interface for using emerge (`emerge.MinimizeAPICalls()` etc.) that hides complexity from users.

### **Coherence**

A measure from 0 to 1 indicating how synchronized a swarm is. 0 means completely random, 1 means perfectly synchronized. Also called the Kuramoto order parameter.

### **Convergence**

The process by which agents reach their target synchronization state. Convergence time varies with scale and goal.

### **Coupling**

The influence agents have on each other. Stronger coupling leads to faster synchronization but may cause oscillation.

### **Coupling Strength (K)**

A parameter determining how strongly agents influence their neighbors' behavior. Must exceed critical coupling for synchronization.

### **Critical Coupling**

The minimum coupling strength needed for synchronization to occur. Below this threshold, agents remain unsynchronized.

### **Cycle**

One complete oscillation from phase 0 to 2π. Agents complete cycles at their frequency rate.

## D

### **Decentralization**

The property of having no central coordinator, master, or single point of control. All agents are equal peers.

### **Decision Maker**

The strategy component that decides how an agent should adjust its state based on observations.

### **Disruption**

An unexpected event that perturbs the system from its goal state. Includes agent failures, network partitions, performance degradation, and environmental changes. Emerge automatically detects and recovers from disruptions. See [Disruption Documentation](disruption.md).

### **Distributed**

Spread across multiple nodes or processes, potentially on different machines.

### **Dynamic**

Changing or adaptive, as opposed to static or fixed. Emerge uses dynamic synchronization that adapts to conditions.

## E

### **Emergence**

Global behavior arising from local interactions without central coordination. Synchronization emerges from agent interactions.

### **Emerge Protocol**

The emerge-specific set of rules agents follow to achieve synchronization. Unique to emerge and based on the Kuramoto model. Defines observation, calculation, adjustment, and recovery phases. See [Protocol Documentation](protocol.md) for details.

### **Energy**

A resource constraint that limits how much agents can adjust. Prevents infinite adjustments and ensures stability.

### **Energy Recovery**

The rate at which agents regain energy over time, typically 5 units per second.

### **Event**

A significant occurrence in the system, such as convergence achieved or strategy switch.

## F

### **Fault Tolerance**

The ability to continue operating despite agent failures. Emerge can tolerate up to 50% agent failure.

### **Frequency**

The rate at which an agent's phase advances through its cycle, measured in Hz or cycles per unit time.

### **Frequency Locking**

When agents align their frequencies before synchronizing phases. A two-stage synchronization process.

### **Full Mesh**

A topology where every agent can observe every other agent. Fast but doesn't scale well.

## G

### **Goal**

A high-level objective that determines what kind of synchronization to achieve (e.g., MinimizeAPICalls, DistributeLoad).

### **Goal-Directed**

Oriented toward achieving a specific objective. Emerge pursues goals through multiple pathways.

### **Goroutine**

Go's lightweight thread. Agents can run as goroutines for in-process coordination.

### **Graceful Degradation**

The ability to maintain partial functionality when some components fail.

## H

### **Hierarchical**

Multiple levels of organization, useful for very large swarms (future feature).

### **Huge Scale**

A predefined configuration for 2000 agents, the largest standard scale.

## I

### **In-Process**

Running within the same operating system process, using goroutines rather than network communication.

### **IsConverged()**

API method to check if the swarm has reached its target synchronization state.

## K

### **Kuramoto Model**

The mathematical model emerge is based on, describing synchronization of coupled oscillators.

### **Kuramoto Order Parameter**

The mathematical measure of synchronization, equivalent to coherence.

## L

### **Large Scale**

A predefined configuration for 1000 agents.

### **Latency**

Time delay. Emerge minimizes coordination latency through local interactions.

### **Leader**

Emerge has no leaders - all agents are equal peers (unlike Raft/Paxos).

### **Local Interactions**

Agents only communicate with immediate neighbors, not the entire swarm.

### **Lock-Free**

Operations that don't require locks, enabling true parallel execution.

## M

### **Medium Scale**

A predefined configuration for 200 agents.

### **Message Passing**

Not used by emerge for synchronization. Agents observe state rather than pass messages.

### **Metronome**

Analogy for emerge agents - they provide timing like a metronome provides beat.

## N

### **Natural Frequency**

An agent's inherent oscillation rate before synchronization adjustments.

### **Neighbors**

The subset of agents that a particular agent can observe and be influenced by.

### **Network Topology**

The pattern of connections between agents (full mesh, ring, small world, etc.).

### **Noise**

Random disturbances that can affect synchronization. Emerge is resilient to noise.

## O

### **Oscillation**

Cyclic behavior. Agents oscillate through phases from 0 to 2π.

### **Order Parameter**

Mathematical term for coherence, measuring the degree of synchronization.

## P

### **Partial Synchronization**

When subgroups synchronize internally but not globally. Used for consensus goals.

### **Pattern**

Request patterns (burst, steady, sparse) describe workload behavior, not synchronization.

### **Peer-to-Peer**

Direct agent-to-agent interaction without intermediaries.

### **Phase**

An agent's position in its oscillation cycle, from 0 to 2π radians. Like the position of a clock hand.

### **Phase Difference**

The angular separation between two agents' phases. Zero difference means perfect synchronization.

### **PulseCoupling**

A strategy using strong, discrete adjustments for rapid synchronization.

## Q

### **Quorum**

Not needed in emerge. Unlike consensus algorithms, emerge doesn't require majority agreement.

## R

### **Recovery**

How agents regain energy or how the system handles failures.

### **Resilience**

Ability to maintain operation despite disruptions, failures, or changes.

### **Ring Topology**

Agents connected in a circle, each seeing only immediate neighbors.

## S

### **Scale**

Predefined configurations for different agent counts (Tiny, Small, Medium, Large, Huge).

### **Self-Organization**

The ability to achieve coordination without external control or commands.

### **Small Scale**

A predefined configuration for 50 agents.

### **Small World**

A topology with mostly local connections plus some long-range connections.

### **SPOF (Single Point of Failure)**

A component whose failure breaks the entire system. Emerge has no SPOF.

### **Strategy**

The approach agents use to achieve synchronization (PhaseNudge, FrequencyLock, PulseCoupling, etc.).

### **Stubbornness**

An agent's resistance to change. Higher stubbornness means slower but more stable convergence.

### **Swarm**

A collection of agents working together toward a common goal.

### **Synchronization**

The process of agents coordinating their behavior to act in harmony.

## T

### **Target State**

The desired synchronization level (coherence) for a particular goal.

### **Thundering Herd**

When many clients act simultaneously, causing spikes. Emerge can prevent or create this deliberately.

### **Tiny Scale**

A predefined configuration for 20 agents, useful for testing.

### **Topology**

The network structure defining which agents can observe which others.

## U

### **Update Interval**

How often agents recalculate and adjust their state.

## V

### **Voting**

Not used in emerge. Unlike consensus algorithms, emerge uses continuous adjustment, not discrete voting.

## W

### **Workload**

The application-specific work that wraps around emerge agents. Emerge handles "when," workload handles "what."

### **Worker Pool**

Goroutines that update agents in parallel for efficient processing.

## Symbols

### **θ (theta)**

Mathematical symbol for phase in the Kuramoto model.

### **ω (omega)**

Mathematical symbol for natural frequency in the Kuramoto model.

### **K**

Coupling strength parameter.

### **N**

Number of agents or neighbors.

### **π (pi)**

Half a cycle (180 degrees) in phase measurement.

### **2π**

Full cycle (360 degrees) in phase measurement.

## Common Phrases

### **Achieve Convergence**

Reach the target synchronization state for the current goal.

### **Emergent Behavior**

Complex global patterns arising from simple local rules.

### **Goal-Directed Synchronization**

Synchronization aimed at achieving specific objectives.

### **Local to Global**

How local interactions lead to global coordination.

### **Multiple Pathways**

Different strategies to reach the same goal, providing resilience.

### **No Central Control**

Fully decentralized with no master, leader, or coordinator.

### **Peer-to-Peer Coordination**

Agents coordinate directly with each other as equals.

### **Phase Space**

The mathematical space in which agent phases exist (0 to 2π).

### **Self-Healing**

Automatic recovery from failures without intervention.

## See Also

- [FAQ](faq.md) - Frequently asked questions
- [Concepts](../concepts/agents.md) - Detailed concept explanations
- [Algorithm](algorithm.md) - Mathematical foundations
