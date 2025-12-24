#include <algorithm>
#include <iostream>
#include <random>
#include <vector>

using ll = long long;

const ll nmax = 2e5 + 5;

std::random_device rd;
std::mt19937 gen(rd());

struct Node {
    ll val, sum, sz, prior;
    Node *l, *r;
    Node() {
    }
    Node(ll x) : val(x), sum(x), sz(1), prior(gen()), l(nullptr), r(nullptr) {
    }
};

int size(Node* v) {
    return (v ? v->sz : 0);
}
ll sm(Node* v) {
    return (v ? v->sum : 0ll);
}

void update(Node* v) {
    v->sz = 1 + size(v->l) + size(v->r);
    v->sum = v->val + sm(v->l) + sm(v->r);
}

Node* merge(Node* A, Node* B) {
    if (!A)
        return B;
    if (!B)
        return A;

    if (A->prior > B->prior) {
        A->r = merge(A->r, B);
        update(A);
        return A;
    } else {
        B->l = merge(A, B->l);
        update(B);
        return B;
    }
}

std::pair<Node*, Node*> split(Node* v, int k) {
    if (!v)
        return {nullptr, nullptr};
    if (1 + size(v->l) <= k) {
        auto [A, B] = split(v->r, k - 1 - size(v->l));
        v->r = A;
        update(v);
        return {v, B};
    } else {
        auto [A, B] = split(v->l, k);
        v->l = B;
        update(v);
        return {A, v};
    }
}

void out(Node* v) {
    if (!v)
        return;
    out(v->l);
    std::cout << v->val << " ";
    out(v->r);
}

void erase(Node *v) {
    if (!v)
        return;
    erase(v->l);
    erase(v->r);
    delete v;
}

Node *even = nullptr, *odd = nullptr;

signed main() {
    std::cin.tie(0)->sync_with_stdio(0);
    int n, q, iters = 1;
    while (true) {
        std::cin >> n >> q;
        if (n + q == 0)
            return 0;
        std::cout << "Swapper " << iters++ << ":\n";
        if (even) erase(even);
        if (odd) erase(odd);
        even = nullptr, odd = nullptr;
        std::vector<int> vec(n);
        for (auto& u : vec)
            std::cin >> u;
        for (int i = 0; i < n; ++i) {
            if (i & 1)
                odd = merge(odd, new Node(vec[i]));
            else
                even = merge(even, new Node(vec[i]));
        }

        while (q--) {
            int typ, l, r, len;
            std::cin >> typ >> l >> r;
            len = r - l + 1;
            l--, r--;
            if (typ == 1) {
                std::pair<int, int> q1, q2;
                q1 = {l / 2, l / 2 + len / 2};
                q2 = {(l + 1) / 2, (l + 1) / 2 + len / 2};
                if (l & 1)
                    std::swap(q1, q2);

                auto [A, B] = split(even, q1.second);
                auto [C, D] = split(A, q1.first);

                auto [A2, B2] = split(odd, q2.second);
                auto [C2, D2] = split(A2, q2.first);

                even = merge(C, merge(D2, B));
                odd = merge(C2, merge(D, B2));
            } else {
                std::pair<int, int> q1, q2;
                q1 = {l / 2, l / 2 + (len + 1) / 2};
                q2 = {(l + 1) / 2, (l + 1) / 2 + len / 2};
                if (l & 1)
                    std::swap(q1, q2);

                auto [A, B] = split(even, q1.second);
                auto [C, D] = split(A, q1.first);

                auto [A2, B2] = split(odd, q2.second);
                auto [C2, D2] = split(A2, q2.first);
                std::cout << sm(D) + sm(D2) << "\n";
                even = merge(C, merge(D, B));
                odd = merge(C2, merge(D2, B2));
            }
        }

        std::cout << "\n";
    }
    return 0;
}
