#!/usr/bin/env perl
use Mojolicious::Lite;
use Mojo::Util 'secure_compare';
use POSIX qw/ceil/;
use utf8;

my $config = plugin 'JSONConfig', { file => '/etc/jackiechat/jackiechat.conf' };

use constant MONTH => {
    0  => 'January',
    1  => 'February',
    2  => 'March',
    3  => 'April',
    4  => 'May',
    5  => 'June',
    6  => 'July',
    7  => 'August',
    8  => 'September',
    9  => 'October',
    10 => 'November',
    11 => 'December',   
};


my $dsn = 'dbi:Pg:dbname=' . $config->{database}{pg}{db_name} . ';host=127.0.0.1';
plugin 'Database', {
    dsn => $dsn,
    username => $config->{database}{pg}{username},
    password => $config->{database}{pg}{password},
    helper => 'db',
};

helper id2chatname => sub {
    my ($self, $id) = @_;

    state $chats = {
        '-1001245038837' => 'RKStatus',
        '-338059465'     => 'TestChatBot',
        '-335599048'     => 'JackieChat Daily',
    };

    my $name = $chats->{$id} // $id;
};

helper id2username => sub {
    my ($self, $id) = @_;

    state $id_users_map = $self->db->selectall_hashref(q{
        SELECT id, name FROM users 
        ORDER BY id
    }, 'id');

    return $id_users_map->{$id}{name} // $id;
};

helper count_messages => sub {
    my ($self, $criteria) = @_;

    my $sql = qq{
        SELECT COUNT(*) 
          FROM messages
    };

    my (@binds, @conditions, $wherestr);
    while (my ($k, $v) = each %$criteria) {
        if (!ref $v) {
            push @conditions, "$k = ?";
            push @binds, $v;
        }
        else {
            if (ref $v eq 'ARRAY') {
                my $condstr = $k . " " . join " ", @$v;
                push @conditions, $condstr;
            }
        }
    }

    if (@conditions) {
        $wherestr = 'WHERE ';
        $wherestr .= join q{ AND }, @conditions;
    }
    
    $sql .= $wherestr if $wherestr;

    my ($cnt) = $self->db->selectrow_array($sql, undef, @binds);

    return $cnt;
};

helper get_messages_list => sub {
    my ($self, $criteria, $opts) = @_;

    my $sql = qq{
        SELECT * FROM messages
    };

    my (@binds, @conditions, $wherestr);
    while (my ($k, $v) = each %$criteria) {
        if (!ref $v) {
            push @conditions, "$k = ?";
            push @binds, $v;
        }
        else {
            if (ref $v eq 'ARRAY') {
                my $condstr = $k . " " . join " ", @$v;
                push @conditions, $condstr;
            }
        }
    }

    if (@conditions) {
        $wherestr = 'WHERE ';
        $wherestr .= join q{ AND }, @conditions;
    }
    
    $sql .= $wherestr if $wherestr;
    $sql .= ' order by ' . $opts->{order_by} if $opts->{order_by};
    $sql .= ' limit '    . $opts->{limit} if $opts->{limit};
    $sql .= ' offset '   . $opts->{offset} // 0;

    my $list = $self->db->selectall_arrayref($sql, { Slice => {} }, @binds);

    return $list;
};

helper get_chats_list => sub {
    my $self = shift;

    my $ids = $self->db->selectcol_arrayref(q{
        SELECT DISTINCT(chat_id) 
          FROM messages
    });

    return $ids;
};

app->config(hypnotoad => { listen => ['http://*:8081'] } );

helper get_users_list => sub {
    my $self = shift;

    my $users = $self->db->selectall_arrayref(q{
        SELECT * FROM users ORDER BY id 
    }, { Slice => {} });

    return $users;
};

get '/' => sub { shift->redirect_to('/dashboard') };

get '/dashboard' => sub {
    my $self = shift;

    $self->render('dashboard');
};

get '/users/:username' => sub {
    my $self = shift;

    $self->stash->{api_key} = $self->config->{users}{$self->stash->{username}};

    $self->render('user');
};

get '/messages/chart' => sub {
    my $self = shift;

    my ($sec,$min,$hour,$mday,$mon,$year,$wday,$yday,$isdst) = localtime(time);
    my (@labels, @data);
    for my $n (0..5) {
        my $m = $mon - $n;
        if ($m < 0) {
            $m = 12 + $m;
        }

        push @labels, MONTH->{$m};

        my $date_from = sprintf q[now() - interval '%d day'], $n * 30 + $mday;
        my $date_to   = $n == 0 ? q[now()] : sprintf q[now() - interval '%d day'], ($n-1) * 30 + $mday;   
        my $sql = qq{
            select count(*)
              from messages m 
             where 
                ctime >= $date_from and ctime <= $date_to 
        };

        my ($cnt) = $self->db->selectrow_array($sql);
        push @data, $cnt;
    }

    $self->render(json => {
        labels => [reverse @labels],
        data   => [reverse @data],
    });
};

get '/messages/:id' => sub {
    my $self = shift;

    $self->stash->{message} = $self->db->selectrow_hashref(q{
        select * from messages where id = ?
    }, undef, $self->stash->{id});

    $self->render('message');
};

post '/messages/:id' => sub {
    my $self = shift;

    my $message = $self->db->selectrow_hashref(q{
        select is_success 
           from messages 
           where id = ?

    }, undef, $self->stash->{id});

    return $self->render(text => 'not found') unless $message;
    return $self->render(text => 'not allowed to edit message that has been sent already') 
        if $message->{is_success} == 1;

    $self->db->do(q{
        update messages 
         set message = ?
         where id = ?
    }, undef, $self->req->param('message'), $self->stash->{id});

    $self->redirect_to('/messages/' . $self->stash->{id});
};

get '/messages' => sub {
    my $self = shift;

    my $q = $self->req->params->to_hash();
    my $p  = delete $q->{p} // 1;
    my $today = delete $q->{today};

    my $limit = 100;
    my $offset = ( $p - 1 ) * $limit;
    
    my %db_query_params = %$q;
    $db_query_params{ctime} = [">", "now() - interval '1 day'"] if $today;

    my $cnt = $self->count_messages(\%db_query_params);

    $self->stash->{n_pages} = ceil $cnt / $limit;
    $self->stash->{cnt} = $cnt;
    $self->stash->{current_page} = $p;

    $self->stash->{messages} = $self->get_messages_list(
        \%db_query_params, 
        {
            order_by => 'ctime desc',
            limit    => $limit,
            offset   => $offset,
        }
    );

    $self->req->query_params->remove('p');
    $self->stash->{qs} = $self->req->query_params->to_string;

    $self->render('messages');
};



post '/messages/:id/enqueue' => sub {
    my $self = shift;

    my $message = $self->db->selectrow_hashref(q{
        select *
          from messages
         where id = ?
    }, undef, $self->stash->{id}) or return $self->render(text => 'not found');

    $self->db->do(q{ 
        update messages 
           set is_success = -1, 
           err = 'waiting for delivery' 
         where id = ? 
    }, undef, $self->stash->{id});

    $self->redirect_to('/messages/' . $self->stash->{id});
};

post '/messages/:id/delete' => sub {
    my $self = shift;

    $self->db->do('delete from messages where id = ?', undef, $self->stash->{id});
    $self->redirect_to('/');
};

app->start;

__DATA__

@@ user.html.ep
% layout 'default';
% title 'JackieChat | User';

<hr />

<div class="row">
    <div class="col-xl-12 col-md-24 mb-4">
        <h3><%= $username %> : <%= $api_key %></h3>
    </div>
</div>


@@ dashboard.html.ep
% layout 'default';
% title 'JackieChat | Dashboard';

<hr />
<!-- Page Heading -->
<div class="d-sm-flex align-items-center justify-content-between mb-4">
    <h1 class="h3 mb-0 text-gray-800">Dashboard</h1>
    <a href="#" class="d-none d-sm-inline-block btn btn-sm btn-primary shadow-sm"><i class="fas fa-download fa-sm text-white-50"></i> Generate Report</a>
</div>

<div class="row">
    <!-- Earnings (Monthly) Card Example -->
    
    <div class="col-xl-3 col-md-6 mb-4">
        <div class="card border-left-primary shadow h-100 py-2">
            <div class="card-body">
                <div class="row no-gutters align-items-center">
                    <div class="col mr-2">
                        <div class="text-xs font-weight-bold text-uppercase text-primary mb-1"><a href="/messages" class="text-primary">Total Sent</a></div>
                        <div class="h5 mb-0 font-weight-bold text-gray-800"><%= count_messages({ is_success => ['!=', '-1'] }) %></div>
                    </div>
                    <div class="col-auto">
                         <i class="fas fa-calendar fa-2x text-gray-300"></i>
                    </div>
                </div>
            </div>
        </div>
    </div>

    <!-- Earnings (Monthly) Card Example -->
    <div class="col-xl-3 col-md-6 mb-4">
        <div class="card border-left-success shadow h-100 py-2">
            <div class="card-body">
                <div class="row no-gutters align-items-center">
                    <div class="col mr-2">
                        <div class="text-xs font-weight-bold text-success text-uppercase mb-1"><a href="/messages?is_success=1" class="text-success">Success Sent</a></div>
                        <div class="h5 mb-0 font-weight-bold text-gray-800"><%= count_messages({ is_success => 1 }) %></div>
                    </div>
                    <div class="col-auto">
                        <i class="fas fa-dollar-sign fa-2x text-gray-300"></i>
                    </div>
                </div>
            </div>
        </div>
    </div>

            <!-- Earnings (Monthly) Card Example -->
    <div class="col-xl-3 col-md-6 mb-4">
        <div class="card border-left-info shadow h-100 py-2">
            <div class="card-body">
                <div class="row no-gutters align-items-center">
                    <div class="col mr-2">
                        <div class="text-xs font-weight-bold text-info text-uppercase mb-1"><a href="/messages?today=1" class="text-info">Today</a></div>
                        <div class="row no-gutters align-items-center">
                            <div class="col-auto">
                                <div class="h5 mb-0 mr-3 font-weight-bold text-gray-800"><%= count_messages({ ctime => [">", "now() - interval '1 day'"] }) %></div>
                            </div>
                        </div>
                    </div>
                </div>
                <div class="col-auto">
                    <i class="fas fa-clipboard-list fa-2x text-gray-300"></i>
                </div>
            </div>
        </div>
    </div>

    <!-- Pending Requests Card Example -->
    <div class="col-xl-3 col-md-6 mb-4">
        <div class="card border-left-warning shadow h-100 py-2">
            <div class="card-body">
                <div class="row no-gutters align-items-center">
                    <div class="col mr-2">
                        <div class="text-xs font-weight-bold text-warning text-uppercase mb-1"><a href="/messages?is_success=-1" class="text-warning">Pending Requests</a></div>
                        <div class="h5 mb-0 font-weight-bold text-gray-800"><%= count_messages({ is_success => -1 }) %></div>
                    </div>
                    <div class="col-auto">
                      <i class="fas fa-comments fa-2x text-gray-300"></i>
                    </div>
                </div>
            </div>
        </div>
    </div>
</div>

<div class="row">
    <!-- Area Chart -->
    <div class="col-xl-12 col-md-24 mb-4">
        <div class="card shadow mb-4">
            <!-- Card Header - Dropdown -->
            <div class="card-header py-3 d-flex flex-row align-items-center justify-content-between">
                <h6 class="m-0 font-weight-bold text-primary">Dynamics</h6>
            </div>
            <!-- Card Body -->
            <div class="card-body">
                <div class="chart-area">
                    <canvas id="myAreaChart"></canvas>
                </div>
            </div>
        </div>
    </div>
</div>

<div class="row">
    <div class="col-xl-6 col-md-12 mb-4">
    % my $users = get_users_list();
        <table class="table table-striped">
            <tr>
                <th>username</th>
                <th class="text-right">total sent</th>
                <th class="text-right">success</th>
                <th class="text-right">today</th>
                <th class="text-right">pending</th>
            </tr>
        % for my $user ($users->@*) {
            <tr>
                <td><a href="/messages?user_id=<%= $user->{id} %>"><%= $user->{name} %></a></td>
                <td class="text-right"><a href="/messages?user_id=<%= $user->{id} %>"><%= count_messages({ user_id => $user->{id}, is_success => ['!=', '-1'] }) %></a></td>
                <td class="text-right"><a href="/messages?user_id=<%= $user->{id} %>&is_success=1"><%= count_messages({ user_id => $user->{id}, is_success => 1 }) %></a></td>
                <td class="text-right"><a href="/messages?user_id=<%= $user->{id} %>&today=1"><%= count_messages({ user_id => $user->{id}, ctime => [">", "now() - interval '1 day'"] }) %></a></td>
                <td class="text-right"><a href="/messages?user_id=<%= $user->{id} %>&is_success=-1"><%= count_messages({ user_id => $user->{id}, is_success => -1 }) %></a></td>
            </tr>
        % }
        </table>
    </div>

    <div class="col-xl-6 col-md-12 mb-4">
    % my $chats = get_chats_list();
        <table class="table table-striped">
            <tr>
                <th>chat name</th>
                <th class="text-right">success</th>
                <th class="text-right">today</th>
                <th class="text-right">pending</th>
            </tr>
        % for my $id (@$chats) {
            <tr>
                <td><a href="/messages?chat_id=<%= $id %>"><%= id2chatname($id) %></a></td>
                <td class="text-right"><a href="/messages?chat_id=<%= $id %>&is_success=1"><%= count_messages({ chat_id => $id, is_success => 1 }) %></a></td>
                <td class="text-right"><a href="/messages?chat_id=<%= $id %>&today=1"><%= count_messages({ chat_id => $id, ctime => [">", "now() - interval '1 day'"] }) %></a></td>
                <td class="text-right"><a href="/messages?chat_id=<%= $id %>&is_success=-1"><%= count_messages({ chat_id => $id, is_success => -1 }) %></a></td>
            </tr>
        % }
        </table>
    </div>
</div>

<script>
    $(document).ready(function() {
        var ctx = document.getElementById('myAreaChart').getContext('2d');
        jQuery.ajax({
            //method: 'get',
            url: '/messages/chart',
            success: function(res) {
                var chart = new Chart(ctx, {
                    // The type of chart we want to create
                    type: 'line',

                    // The data for our dataset
                    data: {
                        labels: res.labels,
                        datasets: [{
                            label: 'Messages sent',
                            backgroundColor: 'rgb(0, 88, 122)',
                            borderColor: 'rgb(0, 88, 122)',
                            data: res.data
                        }]
                    },

                    // Configuration options go here
                    options: {}
                });
                
            },
            error: function() {
                alert('Fail!');
            }
        });

        
    });
</script>


@@ messages.html.ep
% layout 'default';
% title 'JackieChat | Message List';

<hr />
<div class="row">
    <div class="col-xl-12 col-md-24 mb-4">
        <nav aria-label="breadcrumb">
            <ol class="breadcrumb">
                <li class="breadcrumb-item"><a href="/dashboard">Dashboard</a></li>
                <li class="breadcrumb-item active" aria-current="page">Messages</li>
            </ol>
        </nav>
    </div>
</div>

<div class="row">
    <div class="col-xl-12 col-md-24 mb-4">
        <b>user</b>: 
        <select name="username" id="username-select">
            <option <%= 'selected="1"' if !param('user_id') %> value="">All</option>
        % for my $usr (get_users_list()->@*) {
            <option value="<%= $usr->{id} %>" <%= 'selected="1"' if $usr->{id} eq param('user_id') %>><%= $usr->{name} %></option>
        % }
        </select>

        <span class="ml-3 mr-3">- and- </span>

        <b>status</b>:
        <select name="is_success" id="is_success-select"> 
        %   my %status_map = ('1' => 'success', '0' => 'fail', '-1' => 'pending');
            <option <%= 'selected="1"' if !defined param('is_success') %> value="">Any</option>
        % while (my ($is_success, $status) = each(%status_map)) {
            <option value="<%= $is_success %>" <%= 'selected="1"' if defined param('is_success') && param('is_success') eq $is_success %>>
                <%= $status %>
            </option>
        % }
        </select>

        <span class="ml-3 mr-3">- and -</span>

        <b>chat</b>:
        <select name="chat_id" id="chat_id-select">
        % my $chats = get_chats_list();
            <option value="" <%= 'selected="1"' if !param('chat_id') %>>All</option>
        % for my $id (@$chats) {
            <option value="<%= $id %>" <%= 'selected="1"' if $id eq param('chat_id') %>><%= id2chatname($id) %></option>
        % }
        </select>

        <label><input type="checkbox" value="1" id="today-check" class="ml-5" <%= "checked" if param('today') %>> <b>only today</b></label>

        <button id="filter-btn" class="btn btn-sm btn-primary ml-5" onclick="applyFilter()">Apply filter</button>
        <button class="btn btn-danger btn-sm" onclick="resetFilter()">Reset</button>
    </div>
</div>

<div class="row">
    <div class="col-xl-12 col-md-24 mb-4">
        <table class="table">
            <tr>
                <th>#</th>
                <th>Author</th>
                <th>Chat ID</th>
                <th>DateTime</th>
                <th>Result</th>
            </tr>
        % for my $msg (@$messages) {
            <tr>
                <td><a href="/messages/<%= $msg->{id} %>"><%= $msg->{id} %></a></td>
                <td><%= id2username($msg->{user_id}) %></td>
                <td><%= id2chatname($msg->{chat_id}) %></td>
                <td><a href="/messages/<%= $msg->{id} %>"><%= $msg->{ctime} %></a></td>
                <td>
                    % my $class = $msg->{is_success} ? 'badge-success' : 'badge-danger';
                    % $class = 'badge-warning' if $msg->{is_success} eq '-1';
                    % my $title = $msg->{is_success} ? 'Success' : 'Error';
                    % $title = 'Pending' if $msg->{is_success} eq '-1';
                    <span class="badge <%= $class %>"><%= $title %></span>
                </td>
            </tr>
        % }
        </table>

    </div>
</div>

<div class="row">
    <div class="col-xl-12 col-md-24 mb-4">

        <nav aria-label="...">
            <ul class="pagination pagination-sm">
            % for my $page_num (1..$n_pages) {
            %   if ($page_num == $current_page) {
                    <li class="page-item active" aria-current="page">
                        <span class="page-link"><%= $page_num %></span>
                    </li>
            %   }
            %   else {
                    <li class="page-item"><a class="page-link" href='<%= $qs ? "/messages?$qs&p=$page_num" : "/messages?p=$page_num" %>'><%= $page_num %></a></li>
            %   }
            % }
            </ul>
        </nav>

    </div>

</div>



<script>
    function applyFilter() {
        var user_id = document.getElementById('username-select');
        console.log("user_id = ", user_id.value);

        var status = document.getElementById('is_success-select');
        console.log("status = ", status.value);

        var chat = document.getElementById('chat_id-select');
        console.log("chat = ", chat.value);

        var today = document.getElementById('today-check');
        console.log("today = ", today.checked);

        var qs = "?today="; qs += today.checked == true ? "1" : "0";
        if (user_id.value !== "") {
            qs += "&user_id=" + user_id.value;
        }

        if (status.value !== "") {
            qs += "&is_success=" + status.value;
        }

        if (chat.value !== "") {
            qs += "&chat_id=" + chat.value;
        }

        document.location = qs;
    }

    function resetFilter() {
        document.location = "/messages"
    }
</script>

@@ message.html.ep
% layout 'default';
% title 'JackieChat | View Message';

% my $color = $message->{is_success} ? 'success' : 'danger';
% $color = 'warning' if $message->{is_success} eq '-1';

<hr />
<div class="row">
    <div class="col-xl-12 col-md-24 mb-4">
        <nav aria-label="breadcrumb">
            <ol class="breadcrumb">
                <li class="breadcrumb-item"><a href="/dashboard">Dashboard</a></li>
                <li class="breadcrumb-item" aria-current="/messages"><a href="/messages">Messages</a></li>
                <li class="breadcrumb-item active">View</li>
            </ol>
        </nav>
    </div>
</div>
<div class="row">
    <div class="col-md-2"></div>
    <div class="col-md-8">
        <h3>Message:</h3>

        <form method="post">
            <textarea name="message" class="border border-<%= $color %> p-3 w-100 text-monospace" rows="12"><%= $message->{message} %></textarea>
            <input type="submit" class="btn btn-primary btn-sm float-right" value="Edit this message">
        </form>
        
        <div class="pt-2 pb-4"><%= id2username($message->{user_id}) %> | <%= $message->{ctime} %> | <%= id2chatname($message->{chat_id}) %></div>

    % if (!$message->{is_success}) {
        <hr class="bg-danger" />
        <h5 class="text-danger">Not sent:</h5>
        <div class="alert alert-danger p-3"><%= $message->{err} || 'Unknown Error' %></div>
        <form method="post" action="/messages/<%= $message->{id} %>/enqueue">
            <button type="submit" class="btn btn-danger float-right btn-sm">Enqueue</button>
        </form>
    % }

    % if ($message->{is_success} eq '-1') {
        <hr class="bg-warning" />
        <h5 class="text-warning">Not sent yet</h5>
        <div class="alert alert-warning p-3"><%= $message->{err} || 'Unknown Error' %></div>
    % }
    </div>
</div>

@@ layouts/default.html.ep
<!doctype html>
<html lang="en">
  <head>
    <!-- Required meta tags -->
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">

    <!-- Bootstrap CSS -->
    <link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/4.3.1/css/bootstrap.min.css" integrity="sha384-ggOyR0iXCbMQv3Xipma34MD+dH/1fQ784/j6cY/iJTQUOhcWr7x9JvoRxT2MZw1T" crossorigin="anonymous">
    <link href="https://fonts.googleapis.com/css?family=Nunito:200,200i,300,300i,400,400i,600,600i,700,700i,800,800i,900,900i" rel="stylesheet">
    
    <!--<script src="https://code.jquery.com/jquery-3.3.1.slim.min.js" integrity="sha384-q8i/X+965DzO0rT7abK41JStQIAqVgRVzpbzo5smXKp4YfRvH+8abtTE1Pi6jizo" crossorigin="anonymous"></script>-->
    %= javascript '/mojo/jquery/jquery.js'
    <title><%= title %></title>
  </head>
  <body>
    <div class="container">
        <%= content %>
    </div>
    
    <!-- Optional JavaScript -->
    <!-- jQuery first, then Popper.js, then Bootstrap JS -->
    
    <script src="https://cdnjs.cloudflare.com/ajax/libs/popper.js/1.14.7/umd/popper.min.js" integrity="sha384-UO2eT0CpHqdSJQ6hJty5KVphtPhzWj9WO1clHTMGa3JDZwrnQq4sF86dIHNDz0W1" crossorigin="anonymous"></script>
    <script src="https://stackpath.bootstrapcdn.com/bootstrap/4.3.1/js/bootstrap.min.js" integrity="sha384-JjSmVgyd0p3pXB1rRibZUAYoIIy6OrQ6VrjIEaFf/nJGzIxFDsf4x0xIM+B07jRM" crossorigin="anonymous"></script>
    <script src="https://cdn.jsdelivr.net/npm/chart.js@2.8.0"></script>
  </body>
</html>
